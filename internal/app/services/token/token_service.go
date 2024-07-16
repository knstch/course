package token

import (
	"context"
	"errors"

	"github.com/golang-jwt/jwt/v5"
	"github.com/knstch/course/internal/app/config"
	courseError "github.com/knstch/course/internal/app/course_error"
)

type TokenService struct {
	tokenManager tokenManager
	userSecret   string
	adminSecret  string
}

// Claims содержит в себе поля, которые хранятся в JWT.
type UserClaims struct {
	jwt.RegisteredClaims
	Iat      int
	Exp      int
	UserID   uint
	Verified bool
}

// Claims содержит в себе типы данных, которые хранятся в JWT.
type AdminClaims struct {
	jwt.RegisteredClaims
	Iat     int
	Exp     int
	AdminId uint
	Role    string
}

type tokenManager interface {
	DisableToken(ctx context.Context, token string) *courseError.CourseError
	DisableAdminToken(ctx context.Context, token *string) *courseError.CourseError
	CheckAdminAccessToken(ctx context.Context, token *string) *courseError.CourseError
	CheckAccessToken(ctx context.Context, token string) *courseError.CourseError
}

func NewTokenService(tokenManager tokenManager, config *config.Config) *TokenService {
	return &TokenService{
		tokenManager,
		config.Secret,
		config.AdminSecret,
	}
}

// DecodeToken используется для декодирования токена, принимает в качестве параметра токен,
// и возвращает данные из токена или ошибку. Если время жизни токена истекло, то меняет его
// статус в БД на available = false.
func (token TokenService) DecodeUserToken(ctx context.Context, tokenString string) (*UserClaims, *courseError.CourseError) {
	claims := &UserClaims{}

	_, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, nil
		}
		return []byte(token.userSecret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			if err := token.tokenManager.DisableToken(ctx, tokenString); err != nil {
				return nil, err
			}
			return nil, courseError.CreateError(err, 11007)
		}
		return nil, courseError.CreateError(err, 11011)
	}

	if claims.UserID == 0 {
		return nil, courseError.CreateError(err, 11007)
	}

	return claims, nil
}

func (token TokenService) DecodeAdminToken(ctx context.Context, tokenString string) (*AdminClaims, *courseError.CourseError) {
	claims := &AdminClaims{}

	_, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, nil
		}
		return []byte(token.adminSecret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			if err := token.tokenManager.DisableAdminToken(ctx, &tokenString); err != nil {
				return nil, err
			}
			return nil, courseError.CreateError(err, 11007)
		}
		return nil, courseError.CreateError(err, 11011)
	}

	return claims, nil
}

// ValidateAdminAccessToken используется для валидации токена администратора. Используется в middleware.
// Если токен не найден в БД или имеет статус available = false, возвращает ошибку.
func (t TokenService) ValidateAdminAccessToken(ctx context.Context, token *string) *courseError.CourseError {
	if err := t.tokenManager.CheckAdminAccessToken(ctx, token); err != nil {
		return err
	}

	return nil
}

func (t TokenService) ValidateAccessToken(ctx context.Context, token *string) *courseError.CourseError {
	if err := t.tokenManager.CheckAccessToken(ctx, *token); err != nil {
		return err
	}

	return nil
}
