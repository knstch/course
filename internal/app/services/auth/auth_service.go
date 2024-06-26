package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis"
	"github.com/golang-jwt/jwt/v5"
	"github.com/knstch/course/internal/app/config"
	courseError "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/app/services/email"
	"github.com/knstch/course/internal/app/validation"
	"github.com/knstch/course/internal/domain/entity"
)

type Authentificater interface {
	RegisterUser(ctx context.Context, email, password string) (*uint, *courseError.CourseError)
	StoreToken(ctx context.Context, token *string, id *uint) *courseError.CourseError
	SignIn(ctx context.Context, email, password string) (*uint, *bool, *courseError.CourseError)
	VerifyEmail(ctx context.Context, userId uint, isEdit bool) *courseError.CourseError
	DisableTokens(ctx context.Context, userId uint) *courseError.CourseError
	DisableToken(ctx context.Context, token string) *courseError.CourseError
	CheckAccessToken(ctx context.Context, token string) *courseError.CourseError
	RecoverPassword(ctx context.Context, email, password string) *courseError.CourseError
}

type AuthService struct {
	Authentificater Authentificater
	secret          string
	redis           *redis.Client
	emailService    *email.EmailService
}

type Claims struct {
	jwt.RegisteredClaims
	Iat      int
	Exp      int
	UserID   uint
	Verified bool
}

var (
	ErrConfirmCodeNotFound    = errors.New("код не найден")
	ErrBadConfirmCode         = errors.New("код подтверждения не найден")
	ErrEmailIsAlreadyVerified = errors.New("почта уже подтвеждена")
)

func NewAuthService(authentificater Authentificater, config *config.Config, client *redis.Client, emailService *email.EmailService) AuthService {
	return AuthService{
		Authentificater: authentificater,
		secret:          config.Secret,
		redis:           client,
		emailService:    emailService,
	}
}

func (auth AuthService) Register(ctx context.Context, credentials *entity.Credentials) (*string, *courseError.CourseError) {
	if err := validation.NewCredentialsToValidate(credentials).Validate(ctx); err != nil {
		return nil, err
	}

	userId, err := auth.Authentificater.RegisterUser(ctx, credentials.Email, credentials.Password)
	if err != nil {
		return nil, err
	}

	token, err := auth.mintJWT(*userId, false)
	if err != nil {
		return nil, err
	}

	if err := auth.Authentificater.StoreToken(ctx, token, userId); err != nil {
		return nil, err
	}

	if err := auth.emailService.SendConfirmCode(userId, &credentials.Email); err != nil {
		return nil, err
	}

	return token, nil
}

func (auth AuthService) sendConfirmEmail(code uint) *courseError.CourseError {
	if code == 1111 {
		return nil
	}
	return nil
}

func (auth AuthService) generateEmailConfirmCode() uint {
	return 1111
}

func (auth AuthService) VerifyEmail(ctx context.Context, code int, userId uint) (*string, *courseError.CourseError) {
	if err := validation.NewConfirmCodeToValidate(code).Validate(ctx); err != nil {
		return nil, err
	}

	codeFromRedis, err := auth.redis.Get(fmt.Sprint(userId)).Result()
	if err != nil {
		return nil, courseError.CreateError(ErrConfirmCodeNotFound, 11004)
	}

	if fmt.Sprint(code) != codeFromRedis {
		return nil, courseError.CreateError(ErrBadConfirmCode, 11003)
	}

	if err := auth.redis.Del(fmt.Sprint(userId)).Err(); err != nil {
		return nil, courseError.CreateError(err, 10033)
	}

	verificationErr := auth.Authentificater.VerifyEmail(ctx, userId, false)
	if verificationErr != nil {
		return nil, verificationErr
	}

	if err := auth.Authentificater.DisableTokens(ctx, userId); err != nil {
		return nil, err
	}

	token, mintError := auth.mintJWT(userId, true)
	if mintError != nil {
		return nil, courseError.CreateError(mintError.Error, 11010)
	}

	if err := auth.Authentificater.StoreToken(ctx, token, &userId); err != nil {
		return nil, err
	}

	return token, nil
}

func (auth AuthService) SendNewCofirmationCode(ctx context.Context) *courseError.CourseError {
	confimCode := auth.generateEmailConfirmCode()

	code, err := auth.redis.Get(fmt.Sprint(ctx.Value("userId").(uint))).Result()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			return courseError.CreateError(ErrConfirmCodeNotFound, 11004)
		}
	}

	if code != "" {
		if err := auth.redis.Del(fmt.Sprint(ctx.Value("userId").(uint))).Err(); err != nil {
			return courseError.CreateError(err, 10033)
		}
	}

	if err := auth.redis.Set(fmt.Sprint(ctx.Value("userId").(uint)), confimCode, 15*time.Minute).Err(); err != nil {
		return courseError.CreateError(err, 10031)
	}

	if err := auth.sendConfirmEmail(confimCode); err != nil {
		return err
	}

	return nil
}

func (auth AuthService) mintJWT(id uint, verified bool) (*string, *courseError.CourseError) {
	timeNow := time.Now()
	authToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iat":      timeNow.Unix(),
		"exp":      timeNow.Add(30 * 24 * time.Hour).Unix(),
		"userId":   id,
		"verified": verified,
	})

	signedAuthToken, err := authToken.SignedString([]byte(auth.secret))
	if err != nil {
		return nil, courseError.CreateError(err, 11010)
	}

	return &signedAuthToken, nil
}

func (auth AuthService) LogIn(ctx context.Context, credentials *entity.Credentials) (*string, *courseError.CourseError) {
	if err := validation.NewSignInCredentials(credentials).Validate(ctx); err != nil {
		return nil, err
	}

	userId, verified, err := auth.Authentificater.SignIn(ctx, credentials.Email, credentials.Password)
	if err != nil {
		return nil, err
	}

	token, err := auth.mintJWT(*userId, *verified)
	if err != nil {
		return nil, err
	}

	if err := auth.Authentificater.StoreToken(ctx, token, userId); err != nil {
		return nil, err
	}

	return token, nil
}

func (auth AuthService) ValidateAccessToken(ctx context.Context, token *string) *courseError.CourseError {
	if err := auth.Authentificater.CheckAccessToken(ctx, *token); err != nil {
		return err
	}

	return nil
}

func (auth AuthService) DecodeToken(ctx context.Context, tokenString string) (*Claims, *courseError.CourseError) {
	claims := &Claims{}

	_, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, nil
		}
		return []byte(auth.secret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			if err := auth.Authentificater.DisableToken(ctx, tokenString); err != nil {
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

func (auth AuthService) SendPasswordRecoverRequest(ctx context.Context, email string) *courseError.CourseError {
	if err := validation.NewEmailToValidate(email).Validate(ctx); err != nil {
		return err
	}

	if err := auth.emailService.SendPasswordRecoverConfirmCode(email); err != nil {
		return err
	}

	return nil
}

func (auth AuthService) RecoverPassword(ctx context.Context, passwordRecover entity.PasswordRecoverCredentials) *courseError.CourseError {
	if err := validation.NewPasswordRecoverCredentialsToValidate(passwordRecover).Validate(ctx); err != nil {
		return err
	}

	codeFromRedis, err := auth.redis.Get(passwordRecover.Email).Result()
	if err != nil {
		return courseError.CreateError(ErrConfirmCodeNotFound, 11004)
	}

	if fmt.Sprint(passwordRecover.Code) != codeFromRedis {
		return courseError.CreateError(ErrBadConfirmCode, 11003)
	}

	if err := auth.redis.Del(passwordRecover.Email).Err(); err != nil {
		return courseError.CreateError(err, 10033)
	}

	if err := auth.Authentificater.RecoverPassword(ctx, passwordRecover.Email, passwordRecover.Password); err != nil {
		return err
	}

	return nil
}
