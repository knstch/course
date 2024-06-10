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

func (storage *Storage) Automigrate() error {
	if err := storage.db.AutoMigrate(
		&dto.User{},
		&dto.Credentials{},
		&dto.Subscription{},
		&dto.AccessToken{},
	); err != nil {
		return err
	}

	return nil
}

func (storage *Storage) RegisterUser(ctx context.Context, email, password string) (*uint, *courseError.CourseError) {
	credentials := dto.CreateNewCredentials()
	if err := storage.db.Where("email = ?", email).First(&credentials).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			credentials.AddEmail(email).AddPassword(password)
			if err := storage.db.Create(&credentials).Error; err != nil {
				return nil, courseError.CreateError(errRegistingUser, 10001)
			}

			subscription := dto.CreateNewSubscription().AddSubscriptionType("basic")
			if err := storage.db.Create(&subscription).Error; err != nil {
				return nil, courseError.CreateError(errRegistingUser, 10001)
			}

			user := dto.CreateNewUser().
				AddCredentialsId(&credentials.ID).
				AddSubscriptionId(subscription.ID).
				SetStatusUnverified()

			if err := storage.db.Create(&user).Error; err != nil {
				return nil, courseError.CreateError(errRegistingUser, 10001)
			}

			return &user.ID, nil
		}
	}

	return nil, courseError.CreateError(errEmailIsBusy, 11001)
}
