package validation

import (
	"context"
	"fmt"
	"regexp"
	"unicode"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	courseerror "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/domain/entity"
)

type CredentialsToValidate struct {
	Email    string
	Password string
}

const (
	emailPattern    = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	passwordPattern = `^(?=.*[A-Za-z])(?=.*\d)[A-Za-z\d]{8,}$`

	errEmailIsNil = "email обязательно"
	errBadEmail   = "email передан неправильно"

	errPasswordIsNil = "пароль обязателен"
	errBadPassword   = "пароль должен содержать как миниум 8 символов и включать в себя как минимум 1 цифру"
)

var (
	emailRegex = regexp.MustCompile(emailPattern)
)

func NewCredentialsToValidate(credentials *entity.Credentials) *CredentialsToValidate {
	return &CredentialsToValidate{
		Email:    credentials.Email,
		Password: credentials.Password,
	}
}

func (cr *CredentialsToValidate) Validate(ctx context.Context) *courseerror.CourseError {
	if err := validation.ValidateStructWithContext(ctx, cr,
		validation.Field(&cr.Email,
			validation.Required.Error(errEmailIsNil),
			validation.Match(emailRegex).Error(errBadEmail),
			validation.RuneLength(5, 40).Error(errBadEmail),
		),
		validation.Field(&cr.Password,
			validation.Required.Error(errPasswordIsNil),
			validation.By(validatePassword(cr.Password)),
		),
	); err != nil {
		return courseerror.CreateError(err, 400)
	}

	return nil
}

func validatePassword(password string) validation.RuleFunc {
	return func(value interface{}) error {
		if len([]rune(password)) < 8 {
			return fmt.Errorf("пароль должен содержать как миниум 8 символов")
		}

		var hasLetter, hasNumber, hasUpperCase bool

		for _, v := range password {
			switch {
			case unicode.IsLetter(v):
				hasLetter = true
				if unicode.IsUpper(v) {
					hasUpperCase = true
				}
			case unicode.IsNumber(v):
				hasNumber = true
			}
		}

		if !hasLetter || !hasNumber || !hasUpperCase {
			return fmt.Errorf("пароль должен содержать как минимум 1 букву, 1 заглавную букву и 1 цифру")
		}

		return nil
	}
}

type UserInfoToValidate struct {
	firstName   string
	surname     string
	phoneNumber string
}

var (
	phoneRegex = regexp.MustCompile(`^\+?\d{10}$`)
)

func NewUserInfoToValidate(userInfo *entity.UserInfo) *UserInfoToValidate {
	return &UserInfoToValidate{
		firstName:   userInfo.FirstName,
		surname:     userInfo.Surname,
		phoneNumber: userInfo.PhoneNumber,
	}
}

func (user *UserInfoToValidate) Validate(ctx context.Context) *courseerror.CourseError {
	if err := validation.ValidateStructWithContext(ctx, user,
		validation.Field(&user.firstName,
			validation.Required.Error("имя не может быть пустым"),
		),
		validation.Field(&user.surname,
			validation.Required.Error("фамилия не может быть пустой"),
		),
		validation.Field(&user.phoneNumber,
			validation.Required.Error("номер телефона не может быть пустым"),
			validation.Match(phoneRegex).Error("номер телефона передан неверно, введите его в фромате 79123456789"),
		),
	); err != nil {
		return courseerror.CreateError(err, 400)
	}

	return nil
}

type SignInCredentials struct {
	Email    string
	Password string
}

func NewSignInCredentials(credentials *entity.Credentials) *SignInCredentials {
	return &SignInCredentials{
		Email:    credentials.Email,
		Password: credentials.Password,
	}
}

func (cr *SignInCredentials) Validate(ctx context.Context) *courseerror.CourseError {
	if err := validation.ValidateStructWithContext(ctx, cr,
		validation.Field(&cr.Email,
			validation.Required.Error(errEmailIsNil),
		),
		validation.Field(&cr.Password,
			validation.Required.Error(errPasswordIsNil),
		),
	); err != nil {
		return courseerror.CreateError(err, 400)
	}

	return nil
}

type PasswordToValidate struct {
	password string
}

func NewPasswordToValidate(password string) *PasswordToValidate {
	return &PasswordToValidate{
		password: password,
	}
}

func (password *PasswordToValidate) ValidatePassword(ctx context.Context) *courseerror.CourseError {
	if err := validation.ValidateStructWithContext(ctx, password,
		validation.Field(&password.password,
			validation.Required.Error(errPasswordIsNil),
			validation.By(validatePassword(password.password)),
		),
	); err != nil {
		return courseerror.CreateError(err, 400)
	}

	return nil
}

type EmailToValidate struct {
	email string
}

func NewEmailToValidate(email string) *EmailToValidate {
	return &EmailToValidate{
		email: email,
	}
}

func (email *EmailToValidate) Validate(ctx context.Context) *courseerror.CourseError {
	if err := validation.ValidateStructWithContext(ctx, email,
		validation.Field(&email.email,
			validation.Required.Error(errEmailIsNil),
			validation.Match(emailRegex).Error(errBadEmail),
		),
	); err != nil {
		return courseerror.CreateError(err, 400)
	}

	return nil
}
