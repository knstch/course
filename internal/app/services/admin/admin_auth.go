package admin

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	courseError "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/app/validation"
	"github.com/knstch/course/internal/domain/entity"
	"github.com/pquerna/otp/totp"
	qrcode "github.com/skip2/go-qrcode"
)

// RegisterAdmin используется для регистрации нового админа, принимает в качестве параметра
// логин и пароль нового администратора, валидирует их, создает ключ для Google Authentificator,
// генерирует QR-код для добавления ключа в приложение, и возвращает QR-код в виде массива байт и ошибку.
func (admin AdminService) RegisterAdmin(ctx context.Context, credentials *entity.AdminCredentials) ([]byte, *courseError.CourseError) {
	if err := validation.NewAdminCredentialsToValidate(credentials).Validate(ctx); err != nil {
		return nil, err
	}

	if err := admin.adminManager.CheckIfAdminCanBeCreated(ctx, credentials.Login); err != nil {
		return nil, err
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Course",
		AccountName: credentials.Login,
	})
	if err != nil {
		return nil, courseError.CreateError(err, 16050)
	}

	if err := admin.adminManager.AddAdmin(ctx, credentials.Login, credentials.Password, credentials.Role, key.Secret()); err != nil {
		return nil, err
	}

	qrCode, courseErr := admin.generateQrCode(key.String())
	if courseErr != nil {
		return nil, courseErr
	}

	return qrCode, nil
}

// generateQrCode используется для генерации QR-кода, в качестве параметра принимает URL
// и возвращает массив байт, содержащий QR-код, и ошибку.
func (admin AdminService) generateQrCode(url string) ([]byte, *courseError.CourseError) {
	qr, err := qrcode.Encode(url, qrcode.Medium, 256)
	if err != nil {
		return nil, courseError.CreateError(err, 16051)
	}
	return qr, nil
}

// ApproveTwoStepAuth используется для подтверждения двойной аутентификации у администратора.
// Принимает логин, пароль и код, и далее валидирует их.
// Если все параметры оказались валидными, подтверждает получение ключа админом. Возвращает ошибку.
func (admin *AdminService) ApproveTwoStepAuth(ctx context.Context, login, password, code string) *courseError.CourseError {
	signInCredentials := entity.Credentials{
		Email:    login,
		Password: password,
	}

	if err := validation.NewSignInCredentials(&signInCredentials).Validate(ctx); err != nil {
		return err
	}

	_, _, err := admin.adminManager.Login(ctx, login, password, code)
	if err != nil {
		return err
	}

	if err := admin.adminManager.EnableTwoStepAuth(ctx, login, code); err != nil {
		return err
	}

	return nil
}

// SignIn используется для логина администраторов, принимает логин, пароль, и код из Autherntificator.
// Метод валидирует параметры, и если они валидны, возвращает подписанный JWT.
func (admin AdminService) SignIn(ctx context.Context, login, password, code string) (*string, *courseError.CourseError) {
	signInCredentials := entity.Credentials{
		Email:    login,
		Password: password,
	}

	if err := validation.NewSignInCredentials(&signInCredentials).Validate(ctx); err != nil {
		return nil, err
	}

	id, role, err := admin.adminManager.Login(ctx, login, password, code)
	if err != nil {
		return nil, err
	}

	token, err := admin.mintJWT(id, role)
	if err != nil {
		return nil, err
	}

	if err := admin.adminManager.StoreAdminAccessToken(ctx, id, token); err != nil {
		return nil, err
	}

	return token, nil
}

// mintJWT используется для минта JWT, принимает ID администратора и его роль, возвращает токен или ошибку.
func (admin AdminService) mintJWT(id *uint, role *string) (*string, *courseError.CourseError) {
	timeNow := time.Now()
	authToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iat":     timeNow.Unix(),
		"exp":     timeNow.Add(3 * 24 * time.Hour).Unix(),
		"adminId": *id,
		"role":    *role,
	})

	signedAuthToken, err := authToken.SignedString([]byte(admin.secret))
	if err != nil {
		return nil, courseError.CreateError(err, 11010)
	}

	return &signedAuthToken, nil
}

// ValidateAdminAccessToken используется для валидации токена администратора. Используется в middleware.
// Если токен не найден в БД или имеет статус available = false, возвращает ошибку.
func (admin AdminService) ValidateAdminAccessToken(ctx context.Context, token *string) *courseError.CourseError {
	if err := admin.adminManager.CheckAdminAccessToken(ctx, token); err != nil {
		return err
	}

	return nil
}

// DecodeToken используется для декодирования токена, принимает в качестве параметра токен и
// парсит его, и возвращает данные из токена или ошибку.
func (admin AdminService) DecodeToken(ctx context.Context, tokenString string) (*Claims, *courseError.CourseError) {
	claims := &Claims{}

	_, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, nil
		}
		return []byte(admin.secret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			if err := admin.adminManager.DisableAdminToken(ctx, &tokenString); err != nil {
				return nil, err
			}
			return nil, courseError.CreateError(err, 11007)
		}
		return nil, courseError.CreateError(err, 11011)
	}

	return claims, nil
}

// ManageAdminPassword используется для изменения пароля администратора. Принимает логин
// и новый пароль администратора, валидирует их, и изменяет пароль в БД. Возвращает ошибку.
func (admin AdminService) ManageAdminPassword(ctx context.Context, credentials *entity.AdminCredentials) *courseError.CourseError {
	if err := validation.NewAdminCredentialsToValidate(credentials).Validate(ctx); err != nil {
		return err
	}

	if err := admin.adminManager.ResetAdminPassword(ctx, credentials.Login, credentials.Password); err != nil {
		return err
	}

	return nil
}

// ManageAdminAuthKey принимает логин администратора в качестве параметра, валидирует его и
// генерирует новый ключ, используемый в Authentificator. Возвращает QR-код в виде массива байт
// или ошибку.
func (admin AdminService) ManageAdminAuthKey(ctx context.Context, login string) ([]byte, *courseError.CourseError) {
	if login == "" {
		return nil, courseError.CreateError(fmt.Errorf("поле login не может быть пустым"), 400)
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Course",
		AccountName: login,
	})
	if err != nil {
		return nil, courseError.CreateError(err, 16050)
	}

	if err := admin.adminManager.ResetAdminsAuthKey(ctx, login, key.Secret()); err != nil {
		return nil, err
	}

	qrCode, courseErr := admin.generateQrCode(key.String())
	if courseErr != nil {
		return nil, courseErr
	}

	return qrCode, nil
}
