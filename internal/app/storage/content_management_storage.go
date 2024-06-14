package storage

import (
	"context"

	courseError "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/domain/dto"
)

func (storage *Storage) CreateCourse(ctx context.Context, name, description, cost, discount, path string) *courseError.CourseError {
	tx := storage.db.WithContext(ctx).Begin()

	course := dto.CreateNewCourse().
		AddName(name).
		AddDescription(description).
		AddCost(cost).
		AddDiscount(discount).
		AddPreviewImg(path)

	if err := tx.Create(&course).Error; err != nil {
		return courseError.CreateError(err, 10001)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return courseError.CreateError(err, 10010)
	}

	return nil
}
