package validation

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	courseerror "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/domain/entity"
)

type PaymentCredentialsToValidate entity.BuyDetails

func NewPaymentCredentialsToValidate(buyDetails *entity.BuyDetails) *PaymentCredentialsToValidate {
	return (*PaymentCredentialsToValidate)(buyDetails)
}

func (credentials *PaymentCredentialsToValidate) Validate(ctx context.Context) *courseerror.CourseError {
	if err := validation.ValidateStructWithContext(ctx, credentials,
		validation.Field(&credentials.CourseId,
			validation.Required.Error(errFieldIsNil),
		),
		validation.Field(&credentials.IsRusCard,
			validation.Required.Error(errFieldIsNil),
		),
	); err != nil {
		return courseerror.CreateError(err, 400)
	}

	return nil
}

type WebsiteToValidate struct {
	url string
}

func ValidateWebsite(ctx context.Context, link string) *courseerror.CourseError {
	website := &WebsiteToValidate{
		url: link,
	}

	if err := validation.ValidateStructWithContext(ctx, website,
		validation.Field(&website.url,
			is.URL.Error(errBadUrl),
			validation.Required.Error(errFieldIsNil),
		),
	); err != nil {
		return courseerror.CreateError(err, 400)
	}

	return nil
}

func ValidateToken(ctx context.Context, token string) *courseerror.CourseError {
	if err := validation.Validate(token,
		validation.Required.Error(errTokenFieldCantBeEmpty),
	); err != nil {
		return courseerror.CreateError(err, 400)
	}

	return nil
}
