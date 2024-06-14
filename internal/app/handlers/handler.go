package handlers

import (
	"errors"
	"net/http"

	"github.com/go-redis/redis"
	"github.com/knstch/course/internal/app/config"
	"github.com/knstch/course/internal/app/services/auth"
	contentmanagement "github.com/knstch/course/internal/app/services/content_management"
	"github.com/knstch/course/internal/app/services/email"
	"github.com/knstch/course/internal/app/services/user"
	usermanagement "github.com/knstch/course/internal/app/services/user_management"
	"github.com/knstch/course/internal/app/storage"
)

type Handlers struct {
	authService              auth.AuthService
	userService              user.UserService
	emailService             *email.EmailService
	userManagementService    usermanagement.UserManagementService
	contentManagementService contentmanagement.ContentManagementServcie
	address                  string
}

func NewHandlers(storage *storage.Storage, config *config.Config, redisClient *redis.Client, client *http.Client) *Handlers {
	emailService := email.NewEmailService(redisClient)
	return &Handlers{
		authService:              auth.NewAuthService(storage, config, redisClient, emailService),
		userService:              user.NewUserService(storage, emailService, redisClient, client, config.CdnApiKey, config.CdnHost),
		userManagementService:    usermanagement.NewUserManagementService(storage),
		contentManagementService: contentmanagement.NewContentManagementServcie(storage, config, client),
		emailService:             emailService,
		address:                  config.Address,
	}
}

var (
	errBrokenJSON = errors.New("запрос передан в неверном формате")
)
