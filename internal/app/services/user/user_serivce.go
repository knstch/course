package user

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-redis/redis"
	courseError "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/app/services/email"
	"github.com/knstch/course/internal/app/validation"
	"github.com/knstch/course/internal/domain/entity"
)

const (
	fileNamePattern = `\.(.+)$`
)

var (
	fileExtRegex = regexp.MustCompile(fileNamePattern)

	errBadFile          = errors.New("загруженный файл имеет неверный формат")
	errFailedAuth       = errors.New("ошибка авторизации, неверный API ключ или отсутствует userId")
	errCdnFailture      = errors.New("ошибка в CDN")
	errCdnNotResponding = errors.New("CDN не отвечает")

	allowedPhotoFormats = map[string]bool{
		"jpeg": true,
		"jpg":  true,
		"png":  true,
		"JPEG": true,
		"JPG":  true,
		"PNG":  true,
	}
)

type Profiler interface {
	FillUserProfile(ctx context.Context, firstName, surname string, phoneNumber int, userId uint) *courseError.CourseError
	ChangePasssword(ctx context.Context, oldPassword, newPassword string, userId uint) *courseError.CourseError
	ChangeEmail(ctx context.Context, newEmail string, userId uint) *courseError.CourseError
	VerifyEmail(ctx context.Context, userId uint, isEdit bool) *courseError.CourseError
	SetPhoto(ctx context.Context, path string) *courseError.CourseError
	RetreiveUserData(ctx context.Context) (*entity.UserData, *courseError.CourseError)
}

type UserService struct {
	Profiler     Profiler
	emailService *email.EmailService
	redis        *redis.Client
	client       *http.Client
	CdnApiKey    string
	CdnHost      string
}

func NewUserService(profiler Profiler, emailService *email.EmailService, redis *redis.Client, client *http.Client, apiKey, cdnHost string) UserService {
	return UserService{
		Profiler:     profiler,
		emailService: emailService,
		redis:        redis,
		client:       client,
		CdnApiKey:    apiKey,
		CdnHost:      cdnHost,
	}
}

var (
	ErrConfirmCodeNotFound = errors.New("код не найден")
	ErrBadConfirmCode      = errors.New("код подтверждения не найден")
)

func (user UserService) FillProfile(ctx context.Context, userInfo *entity.UserInfo, userId uint) *courseError.CourseError {
	if err := validation.NewUserInfoToValidate(userInfo).Validate(ctx); err != nil {
		return err
	}

	trimedPhoneNumber := strings.TrimPrefix(userInfo.PhoneNumber, "+")

	digitsPhoneNumber, _ := strconv.Atoi(trimedPhoneNumber)

	if err := user.Profiler.FillUserProfile(ctx, userInfo.FirstName, userInfo.Surname, digitsPhoneNumber, userId); err != nil {
		return err
	}

	return nil
}

func (user UserService) EditPassword(ctx context.Context, passwords *entity.Passwords) *courseError.CourseError {
	if err := validation.NewPasswordToValidate(passwords.NewPassword).ValidatePassword(ctx); err != nil {
		return err
	}

	if err := user.Profiler.ChangePasssword(ctx, passwords.OldPassword, passwords.NewPassword, ctx.Value("userId").(uint)); err != nil {
		return err
	}

	return nil
}

func (user UserService) EditEmail(ctx context.Context, email entity.NewEmail, userId uint) *courseError.CourseError {
	if err := validation.NewEmailToValidate(email.NewEmail).Validate(ctx); err != nil {
		return err
	}

	if err := user.Profiler.ChangeEmail(ctx, email.NewEmail, userId); err != nil {
		return err
	}

	if err := user.emailService.SendConfirmCode(&userId, &email.NewEmail); err != nil {
		return err
	}

	return nil
}

func (user UserService) ConfirmEditEmail(ctx context.Context, confirmCode *entity.ConfirmCode, userId uint) *courseError.CourseError {
	if err := validation.NewConfirmCodeToValidate(confirmCode.Code).Validate(ctx); err != nil {
		return err
	}

	codeFromRedis, err := user.redis.Get(fmt.Sprint(userId)).Result()
	if err != nil {
		return courseError.CreateError(ErrConfirmCodeNotFound, 11004)
	}

	if fmt.Sprint(confirmCode.Code) != codeFromRedis {
		return courseError.CreateError(ErrBadConfirmCode, 11003)
	}

	if err := user.redis.Del(fmt.Sprint(userId)).Err(); err != nil {
		return courseError.CreateError(err, 10033)
	}

	if err := user.Profiler.VerifyEmail(ctx, userId, true); err != nil {
		return err
	}

	return nil
}

func (user UserService) AddPhoto(ctx context.Context, formFileName *multipart.FileHeader, file *multipart.File) *courseError.CourseError {
	var fileExtention string
	matches := fileExtRegex.FindStringSubmatch(formFileName.Filename)
	if len(matches) > 1 {
		fileExtention = matches[1]
	} else {
		return courseError.CreateError(errBadFile, 11105)
	}

	_, ok := allowedPhotoFormats[fileExtention]
	if !ok {
		return courseError.CreateError(errBadFile, 11105)
	}

	if err := user.SendPhoto(ctx, file, fileExtention); err != nil {
		return err
	}

	return nil
}

func (user UserService) SendPhoto(ctx context.Context, file *multipart.File, fileExtention string) *courseError.CourseError {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	formFile, err := writer.CreateFormFile("file", fmt.Sprintf("file.%v", fileExtention))
	if err != nil {
		return courseError.CreateError(err, 11031)
	}

	_, err = io.Copy(formFile, *file)
	if err != nil {
		return courseError.CreateError(err, 11042)
	}

	writer.Close()

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%v/upload", user.CdnHost), body)
	if err != nil {
		return courseError.CreateError(err, 11040)
	}

	req.Header.Add("API-KEY", user.CdnApiKey)
	req.Header.Add("userId", fmt.Sprint(ctx.Value("userId").(uint)))
	req.Header.Add("Content-Type", writer.FormDataContentType())

	resp, err := user.client.Do(req)
	if err != nil {
		return courseError.CreateError(errCdnNotResponding, 11041)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return courseError.CreateError(err, 11042)
	}

	cdnResponse := entity.NewCdnResponse()
	if err := json.Unmarshal(respBody, &cdnResponse); err != nil {
		return courseError.CreateError(err, 10101)
	}

	if cdnResponse.Err != nil {
		if cdnResponse.Code == 403 {
			return courseError.CreateError(errFailedAuth, 11050)
		}
		if cdnResponse.Code == 400 {
			return courseError.CreateError(errBadFile, 11105)
		}
		if cdnResponse.Code == 1000 {
			return courseError.CreateError(errCdnFailture, 11051)
		}
	}

	if err := user.Profiler.SetPhoto(ctx, cdnResponse.Path); err != nil {
		return err
	}

	return nil
}

func (user UserService) GetUserInfo(ctx context.Context) (*entity.UserData, *courseError.CourseError) {
	userData, err := user.Profiler.RetreiveUserData(ctx)
	if err != nil {
		return nil, err
	}

	return userData, nil
}
