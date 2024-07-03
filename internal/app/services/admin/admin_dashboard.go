package admin

import (
	"context"
	"time"

	courseError "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/app/validation"
	"github.com/knstch/course/internal/domain/entity"
)

func (admin AdminService) GetPaymentData(ctx context.Context, from, due, courseName, paymentMethod string) ([]entity.PaymentStats, *courseError.CourseError) {
	if err := validation.CreateNewPaymentsQueryToValidate(from, due, courseName, paymentMethod).Validate(ctx); err != nil {
		return nil, err
	}

	dueDate := time.Now()

	fromDate, _ := time.Parse(time.DateOnly, from)

	if due != "" {
		dueDate, _ = time.Parse(time.DateOnly, due)
	}

	stats, err := admin.adminManager.GetSalesStats(ctx, fromDate, dueDate, courseName, paymentMethod)
	if err != nil {
		return nil, err
	}

	return stats, nil
}
