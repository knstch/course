// auth содержит методы для аутентификации, подтверждения почты и смены пароля.
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

// authentificater содержит методы аутентификации для работы с БД.
type authentificater interface {
	RegisterUser(ctx context.Context, email, password string) (*uint, *courseError.CourseError)
	StoreToken(ctx context.Context, token *string, id *uint) *courseError.CourseError
	SignIn(ctx context.Context, email, password string) (*uint, *bool, *courseError.CourseError)
	VerifyEmail(ctx context.Context, userId uint, isEdit bool) *courseError.CourseError
	DisableTokens(ctx context.Context, userId uint) *courseError.CourseError
	DisableToken(ctx context.Context, token string) *courseError.CourseError
	CheckAccessToken(ctx context.Context, token string) *courseError.CourseError
	RecoverPassword(ctx context.Context, email, password string) *courseError.CourseError
}

// AuthService объединяет в себе методы для работы с аутентификацией.
// Содержит в себе redis клиент, email сервис, ключ для подписи
// и декодированию токена и интерфей для взаимодействия с БД.
type AuthService struct {
	authentificater authentificater
	secret          string
	redis           *redis.Client
	emailService    *email.EmailService
}

// Claims содержит в себе поля, которые хранятся в JWT.
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

// NewAuthService - это билдер для сервиса аутентификации.
func NewAuthService(authentificater authentificater, config *config.Config, client *redis.Client, emailService *email.EmailService) AuthService {
	return AuthService{
		authentificater: authentificater,
		secret:          config.Secret,
		redis:           client,
		emailService:    emailService,
	}
}

// Register используется для регистрации нового пользователя. Принимает в качестве
// параметра логин + пароль, валидирует их, регистрирует пользователя, минтит JWT и
// отправляет код подтверждения на почту пользователя. Возвращает JWT и ошибку.
func (auth AuthService) Register(ctx context.Context, credentials *entity.Credentials) (*string, *courseError.CourseError) {
	if err := validation.NewCredentialsToValidate(credentials).Validate(ctx); err != nil {
		return nil, err
	}

	userId, err := auth.authentificater.RegisterUser(ctx, credentials.Email, credentials.Password)
	if err != nil {
		return nil, err
	}

	token, err := auth.mintJWT(*userId, false)
	if err != nil {
		return nil, err
	}

	if err := auth.authentificater.StoreToken(ctx, token, userId); err != nil {
		return nil, err
	}

	if err := auth.emailService.SendConfirmCode(userId, &credentials.Email); err != nil {
		return nil, err
	}

	return token, nil
}

// VerifyEmail используется для верификации почты. Принимает код и ID пользователя в качестве параметров.
// Далее валидируется код, проверяется наличия кода по ID в Redis, если код не совпал, то возвращается ошибка.
// После этого запись удаляется из Redis, пользователь получает статус verified и новый JWT.
// Метод также используется при смене почты, поэтому все другие токены пользователя будут отключены.
func (auth AuthService) VerifyEmail(ctx context.Context, code string, userId uint) (*string, *courseError.CourseError) {
	if err := validation.NewConfirmCodeToValidate(code).Validate(ctx); err != nil {
		return nil, err
	}

	codeFromRedis, err := auth.redis.Get(fmt.Sprint(userId)).Result()
	if err != nil {
		return nil, courseError.CreateError(ErrConfirmCodeNotFound, 11004)
	}

	if code != codeFromRedis {
		return nil, courseError.CreateError(ErrBadConfirmCode, 11003)
	}

	if err := auth.redis.Del(fmt.Sprint(userId)).Err(); err != nil {
		return nil, courseError.CreateError(err, 10033)
	}

	verificationErr := auth.authentificater.VerifyEmail(ctx, userId, false)
	if verificationErr != nil {
		return nil, verificationErr
	}

	if err := auth.authentificater.DisableTokens(ctx, userId); err != nil {
		return nil, err
	}

	token, mintError := auth.mintJWT(userId, true)
	if mintError != nil {
		return nil, courseError.CreateError(mintError.Error, 11010)
	}

	if err := auth.authentificater.StoreToken(ctx, token, &userId); err != nil {
		return nil, err
	}

	return token, nil
}

// SendNewCofirmationCode используется для отправки нового кода на почту пользователя.
// В качестве параметра принимает email пользователя, валидирует его, ищет запись в Redis,
// если она была найдена, то удаляет ее, и отправляет новый код. Возвращает ошибку.
func (auth AuthService) SendNewCofirmationCode(ctx context.Context, email string) *courseError.CourseError {
	if err := validation.NewEmailToValidate(email).Validate(ctx); err != nil {
		return err
	}

	userId := ctx.Value("userId").(uint)

	code, err := auth.redis.Get(fmt.Sprint(userId)).Result()
	if !errors.Is(err, redis.Nil) {
		return courseError.CreateError(ErrConfirmCodeNotFound, 11004)
	}

	if code != "" {
		if err := auth.redis.Del(fmt.Sprint(userId)).Err(); err != nil {
			return courseError.CreateError(err, 10033)
		}
	}

	if err := auth.emailService.SendConfirmCode(&userId, &email); err != nil {
		return err
	}

	return nil
}

// mintJWT используется для минта нового токена доступа для пользователя. Возвращает токен и ошибку.
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

// LogIn метод логина, принимает в качестве параметра пару логин + пароль, валидирует их,
// обращается в БД и проверяет валидность. Далее создает новый токен для пользователя и сохраняет его в БД.
// Возвращает JWT или ошибку.
func (auth AuthService) LogIn(ctx context.Context, credentials *entity.Credentials) (*string, *courseError.CourseError) {
	if err := validation.NewSignInCredentials(credentials).Validate(ctx); err != nil {
		return nil, err
	}

	userId, verified, err := auth.authentificater.SignIn(ctx, credentials.Email, credentials.Password)
	if err != nil {
		return nil, err
	}

	token, err := auth.mintJWT(*userId, *verified)
	if err != nil {
		return nil, err
	}

	if err := auth.authentificater.StoreToken(ctx, token, userId); err != nil {
		return nil, err
	}

	return token, nil
}

// ValidateAccessToken проверяет валидность токена доступа в БД, возвращает ошибку.
func (auth AuthService) ValidateAccessToken(ctx context.Context, token *string) *courseError.CourseError {
	if err := auth.authentificater.CheckAccessToken(ctx, *token); err != nil {
		return err
	}

	return nil
}

// DecodeToken используется для декодирования токена, принимает в качестве параметра токен,
// и возвращает данные из токена или ошибку. Если время жизни токена истекло, то меняет его
// статус в БД на available = false.
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
			if err := auth.authentificater.DisableToken(ctx, tokenString); err != nil {
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

// SendPasswordRecoverRequest используется для отправки кода для восстановления пароля на email
// пользователя. В качестве параметра принимает почту, валидирует ее, и отправляет письмо. Возвращает ошибку.
func (auth AuthService) SendPasswordRecoverRequest(ctx context.Context, email string) *courseError.CourseError {
	if err := validation.NewEmailToValidate(email).Validate(ctx); err != nil {
		return err
	}

	if err := auth.emailService.SendPasswordRecoverConfirmCode(email); err != nil {
		return err
	}

	return nil
}

// RecoverPassword используется для восстановления пароля. В качестве параметра принимает код, почту и пароль.
// Далее валидирует их, проверяет наличие кода в Redis по ключу email, если запись найдена, удаляет ее,
// и меняет пароль в БД. Возвращает ошибку.
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

	if err := auth.authentificater.RecoverPassword(ctx, passwordRecover.Email, passwordRecover.Password); err != nil {
		return err
	}

	return nil
}
