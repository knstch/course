package user

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-redis/redis"
	courseError "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/app/services/email"
	"github.com/knstch/course/internal/app/validation"
	"github.com/knstch/course/internal/domain/entity"
)

type Profiler interface {
	FillUserProfile(ctx context.Context, firstName, surname string, phoneNumber int, userId uint) *courseError.CourseError
	ChangePasssword(ctx context.Context, oldPassword, newPassword string, userId uint) *courseError.CourseError
	ChangeEmail(ctx context.Context, newEmail string, userId uint) *courseError.CourseError
	VerifyEmail(ctx context.Context, userId uint, isEdit bool) *courseError.CourseError
}

type UserService struct {
	Profiler     Profiler
	emailService *email.EmailService
	redis        *redis.Client
}

func NewUserService(profiler Profiler, emailService *email.EmailService, redis *redis.Client) UserService {
	return UserService{
		Profiler:     profiler,
		emailService: emailService,
		redis:        redis,
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

func (user UserService) EditEmail(ctx context.Context, email entity.Email, userId uint) *courseError.CourseError {
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
