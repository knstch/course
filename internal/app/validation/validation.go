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
	passwordPattern = `^[a-zA-Z0-9!@#$%^&*()_+\-=\[\]{};:"\\|,.<>\/?]*$`

	errEmailIsNil                 = "email обязательно"
	errBadEmail                   = "email передан неправильно"
	errPasswordIsNil              = "пароль обязателен"
	errBadPassword                = "пароль должен содержать как миниум 8 символов и включать в себя как минимум 1 цифру"
	errBadConfirmCode             = "код верификации передан неверно"
	errPasswordContainsBadSymbols = "пароль может содержать только латинские буквы, спец. символы и цифры"
)

var (
	emailRegex    = regexp.MustCompile(emailPattern)
	passwordRegex = regexp.MustCompile(passwordPattern)
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
			validation.Match(passwordRegex).Error(errPasswordContainsBadSymbols),
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
	phoneRegex = regexp.MustCompile(`^\+?\d{1,20}$`)
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
			validation.RuneLength(1, 20).Error("имя передано в неверном формате"),
		),
		validation.Field(&user.surname,
			validation.Required.Error("фамилия не может быть пустой"),
			validation.RuneLength(1, 20).Error("фамилия передана в неверном формате"),
		),
		validation.Field(&user.phoneNumber,
			validation.Required.Error("номер телефона не может быть пустым"),
			validation.Match(phoneRegex).Error("номер телефона передан неверно, введите его в фромате 79123456789"),
			validation.RuneLength(1, 20).Error("номер телефона передан в неверном формате"),
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
			validation.Match(passwordRegex).Error(errPasswordContainsBadSymbols),
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
			validation.RuneLength(5, 40).Error(errBadEmail),
		),
	); err != nil {
		return courseerror.CreateError(err, 400)
	}

	return nil
}

type ConfirmCodeToValidate struct {
	code int
}

func NewConfirmCodeToValidate(code int) *ConfirmCodeToValidate {
	return &ConfirmCodeToValidate{
		code: code,
	}
}

func (code *ConfirmCodeToValidate) Validate(ctx context.Context) *courseerror.CourseError {
	if err := validation.ValidateStructWithContext(ctx, code,
		validation.Field(&code.code,
			validation.Min(1000).Error(errBadConfirmCode),
			validation.Max(9999).Error(errBadConfirmCode),
		),
	); err != nil {
		return courseerror.CreateError(err, 400)
	}

	return nil
}

type PasswordRecoverCredentials struct {
	Email    string
	Password string
	Code     int
}

func NewPasswordRecoverCredentialsToValidate(credentials entity.PasswordRecoverCredentials) *PasswordRecoverCredentials {
	return &PasswordRecoverCredentials{
		Email:    credentials.Email,
		Password: credentials.Password,
		Code:     credentials.Code,
	}
}

func (credentials *PasswordRecoverCredentials) Validate(ctx context.Context) *courseerror.CourseError {
	if err := validation.ValidateStructWithContext(ctx, credentials,
		validation.Field(&credentials.Email,
			validation.Required.Error(errEmailIsNil),
			validation.Match(emailRegex).Error(errBadEmail),
			validation.RuneLength(5, 40).Error(errBadEmail),
		),
		validation.Field(&credentials.Password,
			validation.Required.Error(errPasswordIsNil),
			validation.Match(passwordRegex).Error(errPasswordContainsBadSymbols),
			validation.By(validatePassword(credentials.Password)),
		),
		validation.Field(&credentials.Code,
			validation.Min(1000).Error(errBadConfirmCode),
			validation.Max(9999).Error(errBadConfirmCode),
		),
	); err != nil {
		return courseerror.CreateError(err, 400)
	}

	return nil
}
