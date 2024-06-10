package storage

import (
	"context"
	"errors"

	courseError "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/domain/dto"
	"golang.org/x/crypto/bcrypt"
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
	tx := storage.db.WithContext(ctx).Begin()

	credentials := dto.CreateNewCredentials()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password+storage.secret), bcrypt.DefaultCost)
	if err != nil {
		return nil, courseError.CreateError(err, 11020)
	}

	if err := tx.Where("email = ?", email).First(&credentials).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			credentials.AddEmail(email).AddPassword(string(hashedPassword))
			if err := tx.Create(&credentials).Error; err != nil {
				return nil, courseError.CreateError(errRegistingUser, 10001)
			}

			subscription := dto.CreateNewSubscription().AddSubscriptionType("basic")
			if err := tx.Create(&subscription).Error; err != nil {
				return nil, courseError.CreateError(errRegistingUser, 10001)
			}

			user := dto.CreateNewUser().
				AddCredentialsId(&credentials.ID).
				AddSubscriptionId(&subscription.ID).
				SetStatusUnverified()

			if err := tx.Create(&user).Error; err != nil {
				return nil, courseError.CreateError(errRegistingUser, 10001)
			}

			if err := tx.Commit().Error; err != nil {
				return nil, courseError.CreateError(err, 10010)
			}

			return &user.ID, nil
		}

		return nil, courseError.CreateError(err, 10002)
	}

	return nil, courseError.CreateError(errEmailIsBusy, 11001)
}

func (storage *Storage) StoreToken(ctx context.Context, token *string, id *uint) *courseError.CourseError {
	tx := storage.db.WithContext(ctx).Begin()

	accessToken := dto.CreateNewAccessToken().
		AddToken(token).
		AddUsedId(id).
		SetStatusAvailable()

	if err := tx.Create(&accessToken).Error; err != nil {
		return courseError.CreateError(err, 10001)
	}

	if err := tx.Commit().Error; err != nil {
		return courseError.CreateError(err, 10010)
	}

	return nil
}

func (storage *Storage) SignIn(ctx context.Context, email, password string) (*uint, *string, *courseError.CourseError) {
	tx := storage.db.WithContext(ctx).Begin()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password+storage.secret), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, courseError.CreateError(err, 11020)
	}

	credentials := dto.CreateNewCredentials()
	if err := tx.Where("email = ? AND password = ?", email, hashedPassword).First(&credentials).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, courseError.CreateError(err, 11002)
		}
		return nil, nil, courseError.CreateError(err, 10002)
	}

	user := dto.CreateNewUser()
	if err := tx.Where("credentials_id = ?", credentials.ID).First(&user).Error; err != nil {
		return nil, nil, courseError.CreateError(err, 10002)
	}

	subscription := dto.CreateNewSubscription()
	if err := tx.Where("id = ?", user.SubscriptionId).First(&subscription).Error; err != nil {
		return nil, nil, courseError.CreateError(err, 10002)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, nil, courseError.CreateError(err, 10010)
	}

	return &user.ID, &subscription.SubscriptionType, nil
}
