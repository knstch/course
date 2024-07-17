package validation

import (
	"context"
	"fmt"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	courseerror "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/domain/entity"
)

type AdminCredentialsToValidate entity.AdminCredentials

func NewAdminCredentialsToValidate(credentials *entity.AdminCredentials) *AdminCredentialsToValidate {
	return (*AdminCredentialsToValidate)(credentials)
}

func (admin *AdminCredentialsToValidate) Validate(ctx context.Context) *courseerror.CourseError {
	if err := validation.ValidateStructWithContext(ctx, admin,
		validation.Field(&admin.Login,
			validation.Match(loginRegexp).Error(errBadLogin),
		),
		validation.Field(&admin.Password,
			validation.By(validatePassword(admin.Password)),
		),
		validation.Field(&admin.Role,
			validation.In(rolesInterfaces...).Error(errBadRole),
		),
	); err != nil {
		return courseerror.CreateError(err, 400)
	}

	return nil
}

type StatsQueryToValidate struct {
	from,
	due,
	courseName,
	paymentMethod string
}

func CreateNewStatsQueryToValidate(from, due, courseName, paymentMethod string) *StatsQueryToValidate {
	return &StatsQueryToValidate{
		from,
		due,
		courseName,
		paymentMethod,
	}
}

func (query *StatsQueryToValidate) Validate(ctx context.Context) *courseerror.CourseError {
	if err := validation.ValidateStructWithContext(ctx, query,
		validation.Field(&query.courseName,
			validation.RuneLength(1, 100).Error(errBadLength),
		),
		validation.Field(&query.paymentMethod,
			validation.In(paymentMethodsInterfaces...).Error(errBadPaymentMethodParam),
		),
		validation.Field(&query.from,
			validation.Required.Error(errFieldIsNil),
			validation.Date(time.DateOnly).Error(errBadDate),
		),
		validation.Field(&query.due,
			validation.Date(time.DateOnly).Error(errBadDate),
			validation.By(query.validateDue(query.from, query.due)),
		),
	); err != nil {
		return courseerror.CreateError(err, 400)
	}

	return nil
}

func (query *StatsQueryToValidate) validateDue(from, due string) validation.RuleFunc {
	return func(value interface{}) error {
		if due != "" {
			parsedFrom, err := time.Parse(time.DateOnly, from)
			if err != nil {
				return fmt.Errorf(errBadDate)
			}

			parsedDue, err := time.Parse(time.DateOnly, due)
			if err != nil {
				return fmt.Errorf(errBadDate)
			}

			if parsedDue.Before(parsedFrom) {
				return fmt.Errorf(errDueEarlierThenFrom)
			}
		}

		return nil
	}
}

type AdminQueryToValidate struct {
	login              string
	role               string
	twoStepsAuthStatus string
	page               string
	limit              string
}

func CreateNewAdminQueryToValidate(login, role, auth, page, limit string) *AdminQueryToValidate {
	return &AdminQueryToValidate{
		login,
		role,
		auth,
		page,
		limit,
	}
}

func (admin *AdminQueryToValidate) Validate(ctx context.Context) *courseerror.CourseError {
	if err := validation.ValidateStructWithContext(ctx, admin,
		validation.Field(&admin.login,
			validation.RuneLength(1, 200).Error(errBadLength),
		),
		validation.Field(&admin.role,
			validation.RuneLength(1, 200).Error(errBadLength),
		),
		validation.Field(&admin.twoStepsAuthStatus,
			validation.In(boolsInterfaces...).Error(errBadBool),
		),
		validation.Field(&admin.page,
			validation.By(validatePage(admin.page)),
		),
		validation.Field(&admin.limit,
			validation.By(validateLimit(admin.limit)),
		),
	); err != nil {
		return courseerror.CreateError(err, 400)
	}

	return nil
}

type RoleToValidate struct {
	role string
}

func CreateNewRoleToValidate(role string) *RoleToValidate {
	return &RoleToValidate{
		role,
	}
}

func (role *RoleToValidate) Validate(ctx context.Context) *courseerror.CourseError {
	if err := validation.ValidateStructWithContext(ctx, role,
		validation.Field(&role.role,
			validation.In(rolesInterfaces...).Error(errBadRole),
		),
	); err != nil {
		return courseerror.CreateError(err, 400)
	}

	return nil
}
