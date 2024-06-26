package admin

import (
	"context"

	courseError "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/app/validation"
	"github.com/knstch/course/internal/domain/entity"
	"github.com/pquerna/otp/totp"
	qrcode "github.com/skip2/go-qrcode"
)

type adminManager interface {
	AddAdmin(ctx context.Context, login, password, role, key string) *courseError.CourseError
	CheckIfAdminCanBeCreated(ctx context.Context, login string) *courseError.CourseError
	Login(ctx context.Context, login, password string) *courseError.CourseError
	EnableTwoStepAuth(ctx context.Context, login, code string) *courseError.CourseError
}

type AdminService struct {
	adminManager adminManager
}

func NewAdminService(storage adminManager) AdminService {
	return AdminService{
		adminManager: storage,
	}
}

func (admin *AdminService) RegisterAdmin(ctx context.Context, credentials *entity.AdminCredentials) ([]byte, *courseError.CourseError) {
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

func (admin *AdminService) generateQrCode(url string) ([]byte, *courseError.CourseError) {
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

	if err := admin.adminManager.Login(ctx, login, password); err != nil {
		return err
	}

	if err := admin.adminManager.EnableTwoStepAuth(ctx, login, code); err != nil {
		return err
	}

	return nil
}
