package admin

import (
	"context"
	"time"

	courseError "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/app/validation"
	"github.com/knstch/course/internal/domain/entity"
)

// GetPaymentsData используется для выдачи статистики по платежам по дням.
// В качестве параметров принимает дату "от", "до", название курса и платежный метод.
// Обязательным параметром является только "от". Далее они валидируются, собираются данные в БД
// и возвращается массив, содержащий данные по дням, или ошибка.
func (admin AdminService) GetPaymentsData(ctx context.Context, from, due, courseName, paymentMethod string) ([]entity.PaymentStats, *courseError.CourseError) {
	if err := validation.CreateNewStatsQueryToValidate(from, due, courseName, paymentMethod).Validate(ctx); err != nil {
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

// GetUsersData используется для сбора статистики по новым пользователям. Принимает в качестве параметра
// дату "от" и "до", валидирует их, и собирает статистику в БД. Возвращает статистику по новым пользователям
// по дням или ошибку.
func (admin AdminService) GetUsersData(ctx context.Context, from, due string) ([]entity.UsersStats, *courseError.CourseError) {
	if err := validation.CreateNewStatsQueryToValidate(from, due, "", "").Validate(ctx); err != nil {
		return nil, err
	}

	dueDate := time.Now()

	fromDate, _ := time.Parse(time.DateOnly, from)

	if due != "" {
		dueDate, _ = time.Parse(time.DateOnly, due)
	}

	stats, err := admin.adminManager.GetUsersStats(ctx, fromDate, dueDate)
	if err != nil {
		return nil, err
	}

	return stats, nil
}
