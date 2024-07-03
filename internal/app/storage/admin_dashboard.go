package storage

import (
	"context"
	"fmt"
	"strings"
	"time"

	courseError "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/domain/dto"
	"github.com/knstch/course/internal/domain/entity"
)

func (storage Storage) GetStats(ctx context.Context, from, due time.Time, courseName, paymentMethod string) ([]entity.PaymentStats, *courseError.CourseError) {
	tx := storage.db.WithContext(ctx).Begin()

	baseQuery := "SELECT billings.* FROM billings"
	joins := make([]string, 0, 2)
	whereClauses := make([]string, 0, 3)
	whereClauses = append(whereClauses, "DATE(billings.created_at AT TIME ZONE 'UTC') = ? AND billings.paid = ?")
	params := []interface{}{}

	if courseName != "" {
		joins = append(joins, "JOIN orders ON orders.id = billings.order_id")
		joins = append(joins, "JOIN courses ON courses.id = orders.course_id")
		whereClauses = append(whereClauses, "courses.name = ?")
		params = append(params, courseName)
	}

	if paymentMethod != "" {
		whereClauses = append(whereClauses, "billings.payment_method = ?")
		params = append(params, paymentMethod)
	}

	joinClause := strings.Join(joins, " ")
	whereClause := strings.Join(whereClauses, " AND ")
	fullQuery := fmt.Sprintf("%s %s WHERE %s", baseQuery, joinClause, whereClause)

	duration := due.Sub(from)
	daysLeft := int(duration.Hours() / 24)

	stats := make([]entity.PaymentStats, 0, daysLeft)
	for date := from; date.Before(due); date = date.AddDate(0, 0, 1) {
		var billings []dto.Billing

		iterationParams := append([]interface{}{date.Format(time.DateOnly), true}, params...)
		if err := tx.Raw(fullQuery, iterationParams...).Scan(&billings).Error; err != nil {
			tx.Rollback()
			return nil, courseError.CreateError(err, 10002)
		}
		fmt.Printf("Date: %s, Billings fetched: %d\n", date.Format("2006-01-02"), len(billings))
		dayStat := entity.CreateNewPaymentStats(date, billings)

		stats = append(stats, *dayStat)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, courseError.CreateError(err, 10010)
	}

	return stats, nil
}
