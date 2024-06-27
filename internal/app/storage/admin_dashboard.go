package storage

import (
	"context"
	"time"

	courseError "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/domain/dto"
	"github.com/knstch/course/internal/domain/entity"
)

func (storage Storage) GetStats(ctx context.Context, from, due time.Time, courseName, paymentMethod string) ([]entity.PaymentStats, *courseError.CourseError) {
	tx := storage.db.WithContext(ctx).Begin()

	query := tx.Model(&dto.Billing{})

	if courseName != "" {
		query = query.Joins("JOIN courses ON courses.name = ?").
			Joins("JOIN orders ON orders.course_id = courses.id").
			Where("billings.order_id IN (orders.id)")
	}

	if paymentMethod != "" {
		query = query.Where("billings.payment_method = ?", paymentMethod)
	}

	duration := due.Sub(from)
	daysLeft := int(duration.Hours() / 24)

	stats := make([]entity.PaymentStats, 0, daysLeft)
	for date := from; date.Before(due); date = date.AddDate(0, 0, 1) {
		var billings []dto.Billing
		if err := query.Where("created_at = ?", date.Format("2006-01-02")).Find(&billings).Error; err != nil {
			tx.Rollback()
			return nil, courseError.CreateError(err, 10002)
		}
		dayStat := entity.CreateNewPaymentStats(date, billings)

		stats = append(stats, *dayStat)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, courseError.CreateError(err, 10010)
	}

	return stats, nil
}
