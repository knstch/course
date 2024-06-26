package storage

import (
	"context"
	"errors"

	"github.com/knstch/course/internal/app/config"
	courseError "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/domain/dto"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
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

func (storage *Storage) Automigrate(config *config.Config) error {
	if err := storage.db.AutoMigrate(
		&dto.User{},
		&dto.Credentials{},
		&dto.Course{},
		&dto.AccessToken{},
		&dto.Photo{},
		&dto.Order{},
		&dto.Course{},
		&dto.Lesson{},
		&dto.Billing{},
		&dto.Admin{},
		&dto.AdminAccessToken{},
	); err != nil {
		return err
	}

	if config.SuperAdminLogin != "" && config.SuperAdminPassword != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(config.SuperAdminPassword+storage.secret), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		key, err := totp.Generate(totp.GenerateOpts{
			Issuer:      "Course",
			AccountName: config.SuperAdminLogin,
		})
		if err != nil {
			return err
		}

		admin := dto.CreateNewAdmin(config.SuperAdminLogin, string(hashedPassword), "super_admin", key.Secret(), false)
		if err := storage.db.Where("login = ?", config.SuperAdminLogin).FirstOrCreate(&admin).Error; err != nil {
			return err
		}
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
