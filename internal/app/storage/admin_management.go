package storage

import (
	"context"
	"errors"

	courseError "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/domain/dto"
	"gorm.io/gorm"
)

func (storage Storage) RemoveAdmin(ctx context.Context, login string) *courseError.CourseError {
	tx := storage.db.WithContext(ctx).Begin()

	var admin *dto.Admin
	if err := tx.Where("login = ?", login).First(&admin).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return courseError.CreateError(ErrAdminNotFound, 16002)
		}
		return courseError.CreateError(err, 10002)
	}

	if err := tx.Where("login = ?", login).Delete(&dto.Admin{}).Error; err != nil {
		tx.Rollback()
		return courseError.CreateError(err, 10004)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return courseError.CreateError(err, 10010)
	}

	return nil
}

func (storage Storage) ChangeRole(ctx context.Context, login, role string) *courseError.CourseError {
	tx := storage.db.WithContext(ctx).Begin()

	var admin *dto.Admin
	if err := tx.Where("login = ?", login).First(&admin).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return courseError.CreateError(ErrAdminNotFound, 16002)
		}
		return courseError.CreateError(err, 10002)
	}

	if err := tx.Model(&dto.Admin{}).Where("login = ?", login).Update("role", role).Error; err != nil {
		tx.Rollback()
		return courseError.CreateError(err, 10003)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return courseError.CreateError(err, 10010)
	}

	return nil
}
