// user содержит методы для работы с профилем пользователя.
package user

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-redis/redis"
	courseError "github.com/knstch/course/internal/app/course_error"
	cdnerrors "github.com/knstch/course/internal/app/services/cdn_errors"
	"github.com/knstch/course/internal/app/services/email"
	"github.com/knstch/course/internal/app/validation"
	"github.com/knstch/course/internal/domain/entity"
)

type Profiler interface {
	FillUserProfile(ctx context.Context, firstName, surname string, phoneNumber int, userId string) *courseError.CourseError
	ChangePasssword(ctx context.Context, oldPassword, newPassword string, userId uint) *courseError.CourseError
	ChangeEmail(ctx context.Context, newEmail string, userId uint) *courseError.CourseError
	VerifyEmail(ctx context.Context, userId uint, isEdit bool) *courseError.CourseError
	SetPhoto(ctx context.Context, path string) *courseError.CourseError
	RetreiveUserData(ctx context.Context) (*entity.UserData, *courseError.CourseError)
	DeactivateProfile(ctx context.Context) *courseError.CourseError
	SetWatchedStatus(ctx context.Context, lessonId uint) *courseError.CourseError
}

// UserService используется для менеджмента профиля пользователем.
type UserService struct {
	Profiler     Profiler
	emailService *email.EmailService
	redis        *redis.Client
	client       *http.Client
	CdnApiKey    string
	CdnHost      string
}

// NewUserService - это билдер для UserService.
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

// FillProfile используется для заполнения профиля пользователя. Принимает имя, фамилия, номер телефона, ID пользователя и
// источник правок (админ или пользователь). Валидирует параметры и вносит изменения, возвращает ошибку.
func (user UserService) FillProfile(ctx context.Context, userInfo *entity.UserInfo, userId string, isAdminEdit bool) *courseError.CourseError {
	if err := validation.NewUserInfoToValidate(userInfo).Validate(ctx); err != nil {
		return err
	}

	if isAdminEdit {
		if err := validation.NewStringIdToValidate(userId).Validate(ctx); err != nil {
			return err
		}
	}

	trimedPhoneNumber := strings.TrimPrefix(userInfo.PhoneNumber, "+")

	digitsPhoneNumber, _ := strconv.Atoi(trimedPhoneNumber)

	if err := user.Profiler.FillUserProfile(ctx, userInfo.FirstName, userInfo.Surname, digitsPhoneNumber, userId); err != nil {
		return err
	}

	return nil
}

// EditPassword используется для изменения пароля пользователя, принимает в качестве параметра старый и новый пароль,
// валидирует их и вносит изменения. Возвращает ошибку.
func (user UserService) EditPassword(ctx context.Context, passwords *entity.Passwords) *courseError.CourseError {
	if err := validation.NewPasswordToValidate(passwords.NewPassword).ValidatePassword(ctx); err != nil {
		return err
	}

	if err := user.Profiler.ChangePasssword(ctx, passwords.OldPassword, passwords.NewPassword, ctx.Value("UserId").(uint)); err != nil {
		return err
	}

	return nil
}

// EditEmail используется для изменения почты пользователя. Принимает в качестве параметров новую почту и ID пользователя.
// Возвращает ошибку.
func (user UserService) EditEmail(ctx context.Context, email string, userId uint) *courseError.CourseError {
	if err := validation.NewEmailToValidate(email).Validate(ctx); err != nil {
		return err
	}

	if err := user.Profiler.ChangeEmail(ctx, email, userId); err != nil {
		return err
	}

	if err := user.emailService.SendConfirmCode(&userId, &email); err != nil {
		return err
	}

	return nil
}

// ConfirmEditEmail используется для подтверждения изменения почты. Принимает код подтверждения и ID пользователя.
// Метод валидирует параметры и проверяет наличие кода. Возвращает ошибку.
func (user UserService) ConfirmEditEmail(ctx context.Context, confirmCode string, userId uint) *courseError.CourseError {
	if err := validation.NewConfirmCodeToValidate(confirmCode).Validate(ctx); err != nil {
		return err
	}

	codeFromRedis, err := user.redis.Get(fmt.Sprint(userId)).Result()
	if err != nil {
		return courseError.CreateError(ErrConfirmCodeNotFound, 11004)
	}

	if confirmCode != codeFromRedis {
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

// AddPhoto используется для добавления фото. Принимает в качестве параметров фото, валидирует его расширение, загружает
// на CDN и устанавливает фото профиля. Возвращает ошибку.
func (user UserService) AddPhoto(ctx context.Context, formFileHeader *multipart.FileHeader, file *multipart.File) *courseError.CourseError {
	if err := validation.NewImgExtToValidate(formFileHeader.Filename).Validate(ctx); err != nil {
		return err
	}

	path, err := user.sendPhoto(ctx, file, formFileHeader.Filename)
	if err != nil {
		return err
	}

	if err := user.Profiler.SetPhoto(ctx, *path); err != nil {
		return err
	}

	return nil
}

// sendPhoto отправляет фото на CDN и принимает в качестве параметров фото и название название файла.
// Возвращает путь к фото и ошибку.
func (user UserService) sendPhoto(ctx context.Context, file *multipart.File, fileName string) (*string, *courseError.CourseError) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	formFile, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return nil, courseError.CreateError(err, 11031)
	}

	_, err = io.Copy(formFile, *file)
	if err != nil {
		return nil, courseError.CreateError(err, 11042)
	}

	if err := writer.Close(); err != nil {
		return nil, courseError.CreateError(err, 500)
	}

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%v/uploadUserPhoto", user.CdnHost), body)
	if err != nil {
		return nil, courseError.CreateError(err, 11040)
	}

	req.Header.Add("Api-Key", user.CdnApiKey)
	req.Header.Add("UserId", fmt.Sprint(ctx.Value("UserId").(uint)))
	req.Header.Add("Content-Type", writer.FormDataContentType())

	resp, err := user.client.Do(req)
	if err != nil {
		return nil, courseError.CreateError(cdnerrors.ErrCdnNotResponding, 11041)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("ошибка при закрытии тела запроса: %v", err)
		}
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, courseError.CreateError(err, 11042)
	}

	cdnResponse := entity.NewCdnResponse()
	if err := json.Unmarshal(respBody, &cdnResponse); err != nil {
		return nil, courseError.CreateError(err, 10101)
	}

	if cdnResponse.Err != nil {
		if cdnResponse.Code == 403 {
			return nil, courseError.CreateError(cdnerrors.ErrFailedAuth, 11050)
		}
		if cdnResponse.Code == 400 {
			return nil, courseError.CreateError(cdnerrors.ErrBadFile, 11105)
		}
		if cdnResponse.Code == 1000 {
			return nil, courseError.CreateError(cdnerrors.ErrCdnFailture, 11051)
		}
	}

	return &cdnResponse.Path, nil
}

// GetUserInfo используется для получении информации о пользователе. Возвращает даннные пользователя или ошибку.
func (user UserService) GetUserInfo(ctx context.Context) (*entity.UserData, *courseError.CourseError) {
	userData, err := user.Profiler.RetreiveUserData(ctx)
	if err != nil {
		return nil, err
	}

	return userData, nil
}

// DisableProfile используется для деактивации профиля. Возвращает ошибку.
func (user UserService) DisableProfile(ctx context.Context) *courseError.CourseError {
	if err := user.Profiler.DeactivateProfile(ctx); err != nil {
		return err
	}

	return nil
}

// MarkLessonAsWatched используется для отметки пройденного материала. В качестве обязательного параметра
// принимает ID урока, валидирует его и добавляет статус "пройдено" к материалу. Возвращает ошибку.
func (user UserService) MarkLessonAsWatched(ctx context.Context, lessonId string) *courseError.CourseError {
	if err := validation.NewStringIdToValidate(lessonId).Validate(ctx); err != nil {
		return err
	}

	id, _ := strconv.Atoi(lessonId)

	if err := user.Profiler.SetWatchedStatus(ctx, uint(id)); err != nil {
		return err
	}

	return nil
}
