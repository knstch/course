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
			credentials.AddEmail(email).
				AddPassword(string(hashedPassword)).
				SetStatusUnverified()
			if err := tx.Create(&credentials).Error; err != nil {
				tx.Rollback()
				return nil, courseError.CreateError(errRegistingUser, 10001)
			}

			subscription := dto.CreateNewSubscription().AddSubscriptionType("basic")
			if err := tx.Create(&subscription).Error; err != nil {
				tx.Rollback()
				return nil, courseError.CreateError(errRegistingUser, 10001)
			}

			user := dto.CreateNewUser().
				AddCredentialsId(&credentials.ID).
				AddSubscriptionId(&subscription.ID)

			if err := tx.Create(&user).Error; err != nil {
				tx.Rollback()
				return nil, courseError.CreateError(errRegistingUser, 10001)
			}

			if err := tx.Commit().Error; err != nil {
				tx.Rollback()
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
		tx.Rollback()
		return courseError.CreateError(err, 10001)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return courseError.CreateError(err, 10010)
	}

	return nil
}

func (storage *Storage) SignIn(ctx context.Context, email, password string) (userId *uint, subType *string, verified *bool, err *courseError.CourseError) {
	tx := storage.db.WithContext(ctx).Begin()

	credentials := dto.CreateNewCredentials()
	if err := tx.Where("email = ?", email).First(&credentials).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, nil, courseError.CreateError(err, 11002)
		}
		return nil, nil, nil, courseError.CreateError(err, 10002)
	}

	if !storage.verifyPassword(credentials.Password, password) {
		tx.Rollback()
		return nil, nil, nil, courseError.CreateError(errUserNotFound, 11002)
	}

	user := dto.CreateNewUser()
	if err := tx.Where("credentials_id = ?", credentials.ID).First(&user).Error; err != nil {
		tx.Rollback()
		return nil, nil, nil, courseError.CreateError(err, 10002)
	}

	subscription := dto.CreateNewSubscription()
	if err := tx.Where("id = ?", user.SubscriptionId).First(&subscription).Error; err != nil {
		tx.Rollback()
		return nil, nil, nil, courseError.CreateError(err, 10002)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, nil, nil, courseError.CreateError(err, 10010)
	}

	return &user.ID, &subscription.SubscriptionType, &credentials.Verified, nil
}

func (storage *Storage) verifyPassword(hashedPassword, password string) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password+storage.secret)); err != nil {
		return false
	}

	return true
}

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

func (storage *Storage) VerifyUser(ctx context.Context, userId uint) (*string, *courseError.CourseError) {
	tx := storage.db.WithContext(ctx).Begin()

	subscription := dto.CreateNewSubscription()

	if err := tx.Joins("JOIN users ON users.id = ?", userId).
		Where("subscriptions.id = users.subscription_id").
		First(&subscription).Error; err != nil {
		tx.Rollback()
		return nil, courseError.CreateError(err, 11002)
	}

	if err := tx.Exec(`UPDATE "credentials" SET "verified" = ?
		WHERE credentials.id = (SELECT credentials_id 
		FROM "users" WHERE id = ?)`, true, userId).Error; err != nil {
		tx.Rollback()
		return nil, courseError.CreateError(err, 11002)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, courseError.CreateError(err, 10010)
	}

	return &subscription.SubscriptionType, nil
}

func (storage *Storage) DisableTokens(ctx context.Context, userId uint) *courseError.CourseError {
	if err := storage.db.WithContext(ctx).
		Table("access_tokens").
		Where("user_id = ?", userId).
		Update("available", false).Error; err != nil {
		return courseError.CreateError(err, 10003)
	}

	return nil
}

func (storage *Storage) DisableToken(ctx context.Context, token string) *courseError.CourseError {
	if err := storage.db.WithContext(ctx).
		Table("access_tokens").
		Where("token = ?", token).Update("available", false).Error; err != nil {
		return courseError.CreateError(err, 10003)
	}

	return nil
}

func (storage *Storage) CheckAccessToken(ctx context.Context, token string) *courseError.CourseError {
	accessToken := dto.CreateNewAccessToken()

	if err := storage.db.WithContext(ctx).Where("token = ?", token).First(&accessToken).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return courseError.CreateError(errTokenNotFound, 11006)
		}
		return courseError.CreateError(err, 10002)
	}

	return nil
}
