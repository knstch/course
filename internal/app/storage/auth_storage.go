package storage

import (
	"context"
	"errors"

	courseError "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/domain/dto"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

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

func (storage *Storage) VerifyUser(ctx context.Context, userId uint) *courseError.CourseError {
	tx := storage.db.WithContext(ctx).Begin()

	if err := tx.Exec(`DELETE FROM credentials WHERE verified = ? 
		AND credentials.id = (SELECT credentials_id FROM "users" WHERE id = ?)`, true, userId).Error; err != nil {
		tx.Rollback()
		return courseError.CreateError(err, 10004)
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

	if err := storage.db.WithContext(ctx).Where("token = ? AND available = ?", token, true).First(&accessToken).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return courseError.CreateError(errTokenNotFound, 11006)
		}
		return courseError.CreateError(err, 10002)
	}

	return nil
}
