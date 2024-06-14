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
	errUserInactive = errors.New("пользователь неактивен, обратитесь к администратору")
)

func (storage *Storage) RegisterUser(ctx context.Context, email, password string) (*uint, *courseError.CourseError) {
	tx := storage.db.WithContext(ctx).Begin()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password+storage.secret), bcrypt.DefaultCost)
	if err != nil {
		return nil, courseError.CreateError(err, 11020)
	}

	credentials := dto.CreateNewCredentials().AddEmail(email).
		AddPassword(string(hashedPassword)).
		SetStatusUnverified()

	if err := tx.Create(&credentials).Error; err != nil {
		tx.Rollback()
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return nil, courseError.CreateError(errEmailIsBusy, 11001)
		}
		return nil, courseError.CreateError(errRegistingUser, 10001)
	}

	user := dto.CreateNewUser().
		AddCredentialsId(&credentials.ID)

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

func (storage *Storage) SignIn(ctx context.Context, email, password string) (userId *uint, verified *bool, err *courseError.CourseError) {
	tx := storage.db.WithContext(ctx).Begin()

	credentials := dto.CreateNewCredentials()
	if err := tx.Where("email = ?", email).First(&credentials).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, courseError.CreateError(errUserNotFound, 11002)
		}
		return nil, nil, courseError.CreateError(err, 10002)
	}

	if !storage.verifyPassword(credentials.Password, password) {
		tx.Rollback()
		return nil, nil, courseError.CreateError(errUserNotFound, 11002)
	}

	user := dto.CreateNewUser()
	if err := tx.Where("credentials_id = ?", credentials.ID).First(&user).Error; err != nil {
		tx.Rollback()
		return nil, nil, courseError.CreateError(err, 10002)
	}

	if !user.Active {
		tx.Rollback()
		return nil, nil, courseError.CreateError(errUserInactive, 11010)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, nil, courseError.CreateError(err, 10010)
	}

	return &user.ID, &credentials.Verified, nil
}

func (storage *Storage) verifyPassword(hashedPassword, password string) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password+storage.secret)); err != nil {
		return false
	}

	return true
}

func (storage *Storage) DisableTokens(ctx context.Context, userId uint) *courseError.CourseError {
	if err := storage.db.WithContext(ctx).
		Model(dto.AccessToken{}).
		Where("user_id = ?", userId).
		Update("available", false).Error; err != nil {
		return courseError.CreateError(err, 10003)
	}

	return nil
}

func (storage *Storage) DisableToken(ctx context.Context, token string) *courseError.CourseError {
	if err := storage.db.WithContext(ctx).
		Model(dto.AccessToken{}).
		Where("token = ?", token).Update("available", false).Error; err != nil {
		return courseError.CreateError(err, 10003)
	}

	return nil
}

func (storage *Storage) CheckAccessToken(ctx context.Context, token string) *courseError.CourseError {
	accessToken := dto.CreateNewAccessToken()

	if err := storage.db.WithContext(ctx).Where("token = ? AND available = ?", token, true).First(&accessToken).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return courseError.CreateError(errTokenNotFound, 11006)
		}
		return courseError.CreateError(err, 10002)
	}

	return nil
}

func (storage *Storage) RecoverPassword(ctx context.Context, email, password string) *courseError.CourseError {
	tx := storage.db.WithContext(ctx).Begin()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password+storage.secret), bcrypt.DefaultCost)
	if err != nil {
		return courseError.CreateError(err, 11020)
	}

	if err := tx.Model(&dto.Credentials{}).Where("email = ?", email).Update("password", hashedPassword).Error; err != nil {
		return courseError.CreateError(err, 10003)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return courseError.CreateError(err, 10010)
	}

	return nil
}
