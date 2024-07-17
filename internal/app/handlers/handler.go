package handlers

import (
	"errors"
	"net/http"

	"github.com/go-redis/redis"
	"github.com/knstch/course/internal/app/config"
	"github.com/knstch/course/internal/app/grpc"
	"github.com/knstch/course/internal/app/logger"
	"github.com/knstch/course/internal/app/services/admin"
	"github.com/knstch/course/internal/app/services/auth"
	"github.com/knstch/course/internal/app/services/billing"
	contentmanagement "github.com/knstch/course/internal/app/services/content_management"
	"github.com/knstch/course/internal/app/services/email"
	"github.com/knstch/course/internal/app/services/user"
	usermanagement "github.com/knstch/course/internal/app/services/user_management"
	"github.com/knstch/course/internal/app/storage"
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	StatusCodesCounter *prometheus.CounterVec
}

type Handlers struct {
	authService              auth.AuthService
	userService              user.UserService
	emailService             *email.EmailService
	userManagementService    usermanagement.UserManagementService
	contentManagementService contentmanagement.ContentManagementServcie
	sberBillingService       billing.SberBillingService
	adminService             admin.AdminService
	address                  string
	logger                   logger.Logger
	metrics                  MetricsRecorder
}

type MetricsRecorder interface {
	RecordResponse(statusCode int, method, function string)
}

func NewHandlers(storage *storage.Storage,
	config *config.Config,
	redisClient *redis.Client,
	client *http.Client,
	grpcClient *grpc.GrpcClient,
	logger logger.Logger,
	metrics MetricsRecorder) *Handlers {
	emailService := email.NewEmailService(redisClient, config)
	return &Handlers{
		authService:              auth.NewAuthService(storage, config, redisClient, emailService),
		userService:              user.NewUserService(storage, emailService, redisClient, client, config.CdnApiKey, config.CdnHost),
		userManagementService:    usermanagement.NewUserManagementService(storage),
		contentManagementService: contentmanagement.NewContentManagementServcie(storage, config, client, grpcClient),
		sberBillingService:       billing.NewSberBillingService(config, storage, redisClient),
		adminService:             admin.NewAdminService(storage, config.AdminSecret),
		emailService:             emailService,
		address:                  config.HostAddress,
		logger:                   logger,
		metrics:                  metrics,
	}
}

var (
	errBrokenJSON = errors.New("запрос передан в неверном формате")
)
