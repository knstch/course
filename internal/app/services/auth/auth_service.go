package auth

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/knstch/course/internal/app/config"
	courseError "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/app/validation"
	"github.com/knstch/course/internal/domain/entity"
)

type Authentificater interface {
	RegisterUser(ctx context.Context, email, password string) (*uint, *courseError.CourseError)
	FillUserProfile(ctx context.Context, firstName, surname string, phoneNumber int) *courseError.CourseError // вынести в другой сервис
	SignIn(ctx context.Context, email, password string) *courseError.CourseError
	ChangePasssword(ctx context.Context, oldPassword, newPassword string) *courseError.CourseError
}

type AuthService struct {
	Authentificater Authentificater
	secret          string
}

type Claims struct {
	jwt.RegisteredClaims
	Iat    int
	Exp    int
	UserID string
}

func NewAuthService(authentificater Authentificater, config *config.Config) *AuthService {
	return &AuthService{
		Authentificater: authentificater,
		secret:          config.Secret,
	}
}

func (auth *AuthService) Register(ctx context.Context, credentials *entity.Credentials) (*uint, *courseError.CourseError) {
	if err := validation.NewCredentialsToValidate(credentials).Validate(ctx); err != nil {
		return nil, err
	}

	userId, err := auth.Authentificater.RegisterUser(ctx, credentials.Email, credentials.Password)
	if err != nil {
		return nil, err
	}

	return userId, nil
}

func (auth *AuthService) mintJWT(id uint, subscriptionType string) (*string, *courseError.CourseError) {
	timeNow := time.Now()
	authToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iat":              timeNow.Unix(),
		"exp":              timeNow.Add(30 * 24 * time.Hour).Unix(),
		"userId":           id,
		"subscriptionType": subscriptionType,
	})

	signedAuthToken, err := authToken.SignedString(auth.secret)
	if err != nil {
		return nil, courseError.CreateError(err, 11010)
	}

	return &signedAuthToken, nil
}

func (auth *AuthService) FillProfile(ctx context.Context, userInfo *entity.UserInfo) *courseError.CourseError {
	if err := validation.NewUserInfoToValidate(userInfo).Validate(ctx); err != nil {
		return err
	}

	trimedPhoneNumber := strings.TrimPrefix(userInfo.PhoneNumber, "+")

	digitsPhoneNumber, _ := strconv.Atoi(trimedPhoneNumber)

	if err := auth.Authentificater.FillUserProfile(ctx, userInfo.FirstName, userInfo.Surname, digitsPhoneNumber); err != nil {
		return err
	}

	return nil
}

func (auth *AuthService) LogIn(ctx context.Context, credentials *entity.Credentials) *courseError.CourseError {
	if err := validation.NewSignInCredentials(credentials).Validate(ctx); err != nil {
		return err
	}

	if err := auth.Authentificater.SignIn(ctx, credentials.Email, credentials.Password); err != nil {
		return err
	}

	return nil
}

func (auth *AuthService) EditPassword(ctx context.Context, passwords *entity.Passwords) *courseError.CourseError {
	if err := validation.NewPasswordToValidate(passwords.NewPassword).ValidatePassword(ctx); err != nil {
		return err
	}

	if err := auth.Authentificater.ChangePasssword(ctx, passwords.OldPassword, passwords.NewPassword); err != nil {
		return err
	}

	return nil
}
