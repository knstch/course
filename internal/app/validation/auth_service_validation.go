package validation

import (
	"context"
	"strconv"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	courseerror "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/domain/entity"
)

type CredentialsToValidate entity.Credentials

func NewCredentialsToValidate(credentials *entity.Credentials) *CredentialsToValidate {
	return (*CredentialsToValidate)(credentials)
}

func (cr *CredentialsToValidate) Validate(ctx context.Context) *courseerror.CourseError {
	if err := validation.ValidateStructWithContext(ctx, cr,
		validation.Field(&cr.Email,
			validation.Required.Error(errEmailIsNil),
			is.Email.Error(errBadEmail),
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

type ConfirmCodeToValidate struct {
	Code int
	code string
}

func NewConfirmCodeToValidate(code string) *ConfirmCodeToValidate {
	return &ConfirmCodeToValidate{
		code: code,
	}
}

func (code *ConfirmCodeToValidate) Validate(ctx context.Context) *courseerror.CourseError {
	if err := validation.ValidateStructWithContext(ctx, code,
		validation.Field(&code.code,
			is.Int.Error(errBadConfirmCode),
		),
	); err != nil {
		return courseerror.CreateError(err, 400)
	}

	convertedCode, _ := strconv.Atoi(code.code)
	code.Code = convertedCode

	if err := validation.ValidateStructWithContext(ctx, code,
		validation.Field(&code.Code,
			validation.Min(1000).Error(errBadConfirmCode),
			validation.Max(9999).Error(errBadConfirmCode),
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
			is.Email.Error(errBadEmail),
			validation.RuneLength(5, 40).Error(errBadEmail),
		),
	); err != nil {
		return courseerror.CreateError(err, 400)
	}

	return nil
}

type SignInCredentials entity.Credentials

func NewSignInCredentials(credentials *entity.Credentials) *SignInCredentials {
	return (*SignInCredentials)(credentials)
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

type PasswordRecoverCredentials entity.PasswordRecoverCredentials

func NewPasswordRecoverCredentialsToValidate(credentials entity.PasswordRecoverCredentials) *PasswordRecoverCredentials {
	return (*PasswordRecoverCredentials)(&credentials)
}

func (credentials *PasswordRecoverCredentials) Validate(ctx context.Context) *courseerror.CourseError {
	if err := validation.ValidateStructWithContext(ctx, credentials,
		validation.Field(&credentials.Email,
			validation.Required.Error(errEmailIsNil),
			is.Email.Error(errBadEmail),
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
