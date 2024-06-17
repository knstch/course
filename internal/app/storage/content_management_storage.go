package storage

import (
	"context"
	"errors"

	courseError "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/domain/dto"
	"gorm.io/gorm"
)

var (
	errModuleAlreadyExists = errors.New("модуль с такой позицией или названием уже существует")
)

func (storage *Storage) CreateCourse(ctx context.Context, name, description, cost, discount, path string) (*uint, *courseError.CourseError) {
	tx := storage.db.WithContext(ctx).Begin()

	course := dto.CreateNewCourse().
		AddName(name).
		AddDescription(description).
		AddCost(cost).
		AddDiscount(discount).
		AddPreviewImg(path)

	if err := tx.Create(&course).Error; err != nil {
		return nil, courseError.CreateError(err, 10001)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, courseError.CreateError(err, 10010)
	}

	return &course.ID, nil
}

func (storage *Storage) CreateModule(ctx context.Context, name, description string, position, courseId uint) (*uint, *courseError.CourseError) {
	tx := storage.db.WithContext(ctx).Begin()

	module := dto.CreateNewModule()

	if err := storage.db.Where("course_id = ?", courseId).
		Where("position = ?", position).First(&module).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := storage.db.Where("course_id = ?", courseId).
				Where("name = ?", name).First(&module).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					module.AddCourseId(courseId).
						AddName(name).
						AddDescription(description).
						AddPosition(position)

					if err := storage.db.Create(&module).Error; err != nil {
						tx.Rollback()
						return nil, courseError.CreateError(err, 10001)
					}

					if err := tx.Commit().Error; err != nil {
						tx.Rollback()
						return nil, courseError.CreateError(err, 10010)
					}

					return &module.ID, nil
				}
				tx.Rollback()
				return nil, courseError.CreateError(err, 10002)
			}
		}
		tx.Rollback()
		return nil, courseError.CreateError(errModuleAlreadyExists, 13001)
	}
	tx.Rollback()
	return nil, courseError.CreateError(errModuleAlreadyExists, 13001)
}
