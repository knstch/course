package admin

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	courseError "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/app/validation"
	"github.com/knstch/course/internal/domain/entity"
	"github.com/pquerna/otp/totp"
	qrcode "github.com/skip2/go-qrcode"
)

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

func (admin AdminService) generateQrCode(url string) ([]byte, *courseError.CourseError) {
	qr, err := qrcode.Encode(url, qrcode.Medium, 256)
	if err != nil {
		return nil, courseError.CreateError(err, 16051)
	}
	return qr, nil
}

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

func (admin AdminService) ValidateAdminAccessToken(ctx context.Context, token *string) *courseError.CourseError {
	if err := admin.adminManager.CheckAdminAccessToken(ctx, token); err != nil {
		return err
	}

	return nil
}

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
