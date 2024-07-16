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

	if err := tx.Model(&dto.Admin{}).Where("login = ?", login).Update("Role", role).Error; err != nil {
		tx.Rollback()
		return courseError.CreateError(err, 10003)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return courseError.CreateError(err, 10010)
	}

	return nil
}

func (storage Storage) GetAdmins(ctx context.Context, login, role, auth string, limit, offset int) ([]dto.Admin, *courseError.CourseError) {
	tx := storage.db.WithContext(ctx).Begin()

	query := tx.Model(&dto.Admin{})

	if login != "" {
		query = query.Where("login = ?", login)
	}

	if role != "" {
		query = query.Where("role = ?", role)
	}

	if auth != "" {
		var authStatus bool
		if auth == "true" {
			authStatus = true
		}
		query = query.Where("two_steps_auth_enabled = ?", authStatus)
	}

	var admins []dto.Admin
	if err := query.Offset(offset).Limit(limit).Find(&admins).Error; err != nil {
		return nil, courseError.CreateError(err, 10002)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, courseError.CreateError(err, 10010)
	}

	return admins, nil
}
