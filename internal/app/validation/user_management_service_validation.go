package validation

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	courseerror "github.com/knstch/course/internal/app/course_error"
)

type UserFiltersToValidate struct {
	firstName   string
	surname     string
	phoneNumber string
	active      string
	email       string
	isVerified  string
	page        string
	limit       string
	banned      string
}

func NewUserFiltersToValidate(firstName, surname, phoneNumber, email, active, isVerified, banned, page, limit string) *UserFiltersToValidate {
	return &UserFiltersToValidate{
		firstName:   firstName,
		surname:     surname,
		phoneNumber: phoneNumber,
		active:      active,
		email:       email,
		isVerified:  isVerified,
		page:        page,
		limit:       limit,
		banned:      banned,
	}
}

func (userFilters *UserFiltersToValidate) Validate(ctx context.Context) *courseerror.CourseError {
	if err := validation.ValidateStructWithContext(ctx, userFilters,
		validation.Field(&userFilters.firstName,
			validation.Match(lettersRegex).Error(errFieldAcceptsOnlyLetters),
		),
		validation.Field(&userFilters.surname,
			validation.Match(lettersRegex).Error(errFieldAcceptsOnlyLetters),
		),
		validation.Field(&userFilters.phoneNumber,
			validation.Match(phoneRegex).Error("номер телефона передан неверно, введите его в фромате 79123456789"),
		),
		validation.Field(&userFilters.isVerified,
			validation.In(boolsInterfaces...).Error(errBadBool),
		),
		validation.Field(&userFilters.active,
			validation.In(boolsInterfaces...).Error(errBadBool),
		),
		validation.Field(&userFilters.banned,
			validation.In(boolsInterfaces...).Error(errBadBool),
		),
		validation.Field(&userFilters.email,
			is.Email.Error(errBadEmail),
		),
		validation.Field(&userFilters.page,
			validation.By(validatePage(userFilters.page)),
		),
		validation.Field(&userFilters.limit,
			validation.By(validateLimit(userFilters.limit)),
		),
	); err != nil {
		return courseerror.CreateError(err, 400)
	}

	return nil
}
