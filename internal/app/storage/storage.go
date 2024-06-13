package storage

import (
	"context"
	"errors"

	courseError "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/domain/dto"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	errRegistingUser = errors.New("ошибка при регистрации пользователя")
	errEmailIsBusy   = errors.New("пользователь с таким email уже существует")
	errUserNotFound  = errors.New("пользователь не найден")
	errTokenNotFound = errors.New("токен не найден")
)

type Storage struct {
	db     *gorm.DB
	secret string
}

func NewStorage(dsn, secret string) (*Storage, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return &Storage{
		db:     db,
		secret: secret,
	}, nil
}

func newUserProfileUpdate(firstName, surname string, phoneNumber int) map[string]interface{} {
	updates := make(map[string]interface{}, 3)

	updates["phone_number"] = phoneNumber
	updates["first_name"] = firstName
	updates["surname"] = surname

	return updates
}

func (storage *Storage) Automigrate() error {
	if err := storage.db.AutoMigrate(
		&dto.User{},
		&dto.Credentials{},
		&dto.Course{},
		&dto.AccessToken{},
		&dto.Photo{},
		&dto.UsersCourse{},
		&dto.Course{},
	); err != nil {
		return err
	}

	return nil
}

func (storage *Storage) VerifyEmail(ctx context.Context, userId uint, isEdit bool) *courseError.CourseError {
	tx := storage.db.WithContext(ctx).Begin()

	if isEdit {
		oldCredentials := dto.CreateNewCredentials()

		if err := tx.Joins("JOIN users ON users.id = ?", userId).
			Where("credentials.id = users.credentials_id AND verified = ?", true).
			First(&oldCredentials).Error; err != nil {
			tx.Rollback()
			return courseError.CreateError(err, 11002)
		}

		if err := tx.Exec(`UPDATE "users" SET credentials_id = 
			(SELECT id FROM "credentials" WHERE verified = ? 
			ORDER BY created_at DESC LIMIT 1) WHERE id = ?`, false, userId).Error; err != nil {
			tx.Rollback()
			return courseError.CreateError(err, 11002)
		}

		if err := tx.Where("id = ?", oldCredentials.ID).Delete(&dto.Credentials{}).Error; err != nil {
			tx.Rollback()
			return courseError.CreateError(err, 10004)
		}
	}

	if err := tx.Exec(`UPDATE "credentials" SET "verified" = ?
		WHERE credentials.id = (SELECT credentials_id 
		FROM "users" WHERE id = ?) AND verified = ?`, true, userId, false).Error; err != nil {
		tx.Rollback()
		return courseError.CreateError(err, 11002)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return courseError.CreateError(err, 10010)
	}

	return nil
}
