package storage

import (
	"context"
	"errors"

	courseError "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/domain/dto"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrAdminLoginExits       = errors.New("логин занят")
	ErrAdminNotFound         = errors.New("администратор не найден")
	ErrBadAuthCode           = errors.New("неверный код")
	ErrAdminAccessProhibited = errors.New("неверный логин или пароль")
)

func (storage Storage) AddAdmin(ctx context.Context, login, password, role, key string) *courseError.CourseError {
	tx := storage.db.WithContext(ctx).Begin()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password+storage.secret), bcrypt.DefaultCost)
	if err != nil {
		tx.Rollback()
		return courseError.CreateError(err, 11020)
	}

	admin := dto.CreateNewAdmin(login, string(hashedPassword), role, key, false)
	if err := tx.Create(&admin).Error; err != nil {
		tx.Rollback()
		return courseError.CreateError(err, 10001)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return courseError.CreateError(err, 10010)
	}

	return nil
}

func (storage Storage) CheckIfAdminCanBeCreated(ctx context.Context, login string) *courseError.CourseError {
	tx := storage.db.WithContext(ctx).Begin()

	var admin *dto.Admin
	if err := tx.Where("login = ?", login).First(&admin).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		tx.Rollback()
		return courseError.CreateError(err, 10002)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return courseError.CreateError(err, 10010)
	}

	return courseError.CreateError(ErrAdminLoginExits, 16001)
}

func (storage Storage) EnableTwoStepAuth(ctx context.Context, login, code string) *courseError.CourseError {
	tx := storage.db.WithContext(ctx).Begin()

	var adminCredentials *dto.Admin
	if err := tx.Where("login = ?", login).First(&adminCredentials).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return courseError.CreateError(ErrAdminNotFound, 16002)
		}
		return courseError.CreateError(err, 10002)
	}

	valid := totp.Validate(code, adminCredentials.Key)
	if !valid {
		return courseError.CreateError(ErrBadAuthCode, 16052)
	}

	if err := tx.Model(&dto.Admin{}).Where("login = ?", login).Update("two_steps_auth_enabled", true).Error; err != nil {
		return courseError.CreateError(err, 10003)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return courseError.CreateError(err, 10010)
	}

	return nil
}

func (storage Storage) Login(ctx context.Context, login, password string) *courseError.CourseError {
	tx := storage.db.WithContext(ctx).Begin()

	var credentials *dto.Admin
	if err := tx.Where("login = ?", login).First(&credentials).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return courseError.CreateError(ErrAdminNotFound, 16002)
		}
		return courseError.CreateError(err, 10002)
	}

	verified := storage.verifyPassword(credentials.Password, password)
	if !verified {
		return courseError.CreateError(ErrAdminAccessProhibited, 16003)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return courseError.CreateError(err, 10010)
	}

	return nil
}
