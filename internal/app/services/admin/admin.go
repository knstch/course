// admin содержит методы для администрирования платформы.
package admin

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	courseError "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/domain/dto"
	"github.com/knstch/course/internal/domain/entity"
)

// adminManager объединяет в себе методы администратора по взаимодействию с БД.
type adminManager interface {
	AddAdmin(ctx context.Context, login, password, role, key string) *courseError.CourseError
	CheckIfAdminCanBeCreated(ctx context.Context, login string) *courseError.CourseError
	Login(ctx context.Context, login, password, code string) (*uint, *string, *courseError.CourseError)
	EnableTwoStepAuth(ctx context.Context, login, code string) *courseError.CourseError
	StoreAdminAccessToken(ctx context.Context, id *uint, token *string) *courseError.CourseError
	CheckAdminAccessToken(ctx context.Context, token *string) *courseError.CourseError
	DisableAdminToken(ctx context.Context, token *string) *courseError.CourseError
	RemoveAdmin(ctx context.Context, login string) *courseError.CourseError
	ChangeRole(ctx context.Context, login, role string) *courseError.CourseError
	GetAdmins(ctx context.Context, login, role, auth string, limit, offset int) ([]dto.Admin, *courseError.CourseError)
	ResetAdminPassword(ctx context.Context, login, newPassword string) *courseError.CourseError
	ResetAdminsAuthKey(ctx context.Context, login, key string) *courseError.CourseError
	GetSalesStats(ctx context.Context, from, due time.Time, courseName, paymentMethod string) ([]entity.PaymentStats, *courseError.CourseError)
	GetUsersStats(ctx context.Context, from, due time.Time) ([]entity.UsersStats, *courseError.CourseError)
}

// Claims содержит в себе типы данных, которые хранятся в JWT.
type Claims struct {
	jwt.RegisteredClaims
	Iat     int
	Exp     int
	AdminId uint
	Role    string
}

// AdminService объединяет в себе методы администратора и ключ для шифрования и расшифровки JWT.
type AdminService struct {
	adminManager adminManager
	secret       string
}

// NewAdminService билдер для сервиса администратров.
func NewAdminService(storage adminManager, secret string) AdminService {
	return AdminService{
		adminManager: storage,
		secret:       secret,
	}
}
