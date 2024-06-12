package storage

import (
	"context"
	"errors"

	courseError "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/domain/dto"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	errBadPassword = errors.New("пароль передан неверно")
)

func (storage *Storage) FillUserProfile(ctx context.Context, firstName, surname string, phoneNumber int, userId uint) *courseError.CourseError {
	tx := storage.db.WithContext(ctx).Begin()

	user := dto.CreateNewUser()

	if err := tx.Where("id = ?", userId).First(&user).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return courseError.CreateError(err, 11101)
		}
		return courseError.CreateError(err, 10002)
	}

	if err := tx.Where("id = ?", userId).Updates(newUserProfileUpdate(firstName, surname, phoneNumber)).Error; err != nil {
		tx.Rollback()
		return courseError.CreateError(err, 10003)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return courseError.CreateError(err, 10010)
	}

	return nil
}

func (storage *Storage) ChangePasssword(ctx context.Context, oldPassword, newPassword string, userId uint) *courseError.CourseError {
	credentials := dto.CreateNewCredentials()

	tx := storage.db.WithContext(ctx).Begin()

	if err := tx.Joins("JOIN users ON users.id = ?", userId).
		Where("credentials.id = users.credentials_id").
		First(&credentials).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return courseError.CreateError(errUserNotFound, 11002)
		}
		return courseError.CreateError(err, 10002)
	}

	if !storage.verifyPassword(credentials.Password, oldPassword) {
		return courseError.CreateError(errBadPassword, 11102)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword+storage.secret), bcrypt.DefaultCost)
	if err != nil {
		return courseError.CreateError(err, 11020)
	}

	if err := tx.Where("id = ?", credentials.ID).Update("password", hashedPassword).Error; err != nil {
		return courseError.CreateError(err, 10003)
	}

	return nil
}
