package storage

import (
	"context"
	"errors"
	"strings"

	courseError "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/domain/dto"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	errBadPassword             = errors.New("старый пароль передан неверно")
	errOldAndNewPasswordsEqual = errors.New("новый и старый пароль не могут совпадать")
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

	if err := tx.Model(&dto.User{}).Where("id = ?", userId).Updates(newUserProfileUpdate(firstName, surname, phoneNumber)).Error; err != nil {
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
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return courseError.CreateError(errUserNotFound, 11002)
		}
		return courseError.CreateError(err, 10002)
	}

	if oldPassword == newPassword {
		tx.Rollback()
		return courseError.CreateError(errOldAndNewPasswordsEqual, 11103)
	}

	if !storage.verifyPassword(credentials.Password, oldPassword) {
		tx.Rollback()
		return courseError.CreateError(errBadPassword, 11104)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword+storage.secret), bcrypt.DefaultCost)
	if err != nil {
		tx.Rollback()
		return courseError.CreateError(err, 11020)
	}

	if err := tx.Model(dto.Credentials{}).Where("id = ?", credentials.ID).Update("password", hashedPassword).Error; err != nil {
		tx.Rollback()
		return courseError.CreateError(err, 10003)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return courseError.CreateError(err, 10010)
	}

	return nil
}

func (storage *Storage) ChangeEmail(ctx context.Context, newEmail string, userId uint) *courseError.CourseError {
	tx := storage.db.WithContext(ctx).Begin()

	oldCredentials := dto.CreateNewCredentials()

	if err := tx.Joins("JOIN users ON users.id = ?", userId).
		Where("credentials.id = users.credentials_id AND verified = ?", true).
		First(&oldCredentials).Error; err != nil {
		tx.Rollback()
		return courseError.CreateError(err, 11002)
	}

	newCredentials := dto.CreateNewCredentials().
		AddPassword(oldCredentials.Password).
		AddEmail(newEmail).SetStatusUnverified()

	if err := tx.Create(&newCredentials).Error; err != nil {
		tx.Rollback()
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return courseError.CreateError(errEmailIsBusy, 11001)
		}
		return courseError.CreateError(err, 10001)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return courseError.CreateError(err, 10010)
	}

	return nil
}

func (storage *Storage) SetPhoto(ctx context.Context, path string) *courseError.CourseError {
	tx := storage.db.WithContext(ctx).Begin()

	userId := ctx.Value("userId").(uint)

	photo := dto.CreateNewPhoto(path)
	if err := tx.Create(&photo).Error; err != nil {
		return courseError.CreateError(err, 10001)
	}

	if err := tx.Model(&dto.User{}).Where("id = ?", userId).Update("photo_id", photo.ID).Error; err != nil {
		return courseError.CreateError(err, 10003)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return courseError.CreateError(err, 10010)
	}

	return nil
}
