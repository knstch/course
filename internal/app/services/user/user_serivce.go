package user

import (
	"context"
	"strconv"
	"strings"

	courseError "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/app/validation"
	"github.com/knstch/course/internal/domain/entity"
)

type Profiler interface {
	FillUserProfile(ctx context.Context, firstName, surname string, phoneNumber int, userId uint) *courseError.CourseError
	ChangePasssword(ctx context.Context, oldPassword, newPassword string, userId uint) *courseError.CourseError
	// ChangeEmail(ctx context.Context, newEmail string) *courseError.CourseError
}

type UserService struct {
	Profiler Profiler
}

func NewUserService(profiler Profiler) UserService {
	return UserService{
		Profiler: profiler,
	}
}

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

// func (user UserService) EditEmail(ctx context.Context, emails entity.Emails) *courseError.CourseError {
// 	if err := validation.NewEmailToValidate(emails.NewEmail).Validate(ctx); err != nil {
// 		return err
// 	}

// 	if err := user.Profiler.ChangeEmail(ctx, emails.NewEmail); err != nil {
// 		return err
// 	}

// 	return nil
// }
