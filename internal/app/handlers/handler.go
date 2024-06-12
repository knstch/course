package handlers

import (
	"errors"

	"github.com/go-redis/redis"
	"github.com/knstch/course/internal/app/config"
	"github.com/knstch/course/internal/app/services/auth"
	"github.com/knstch/course/internal/app/services/email"
	"github.com/knstch/course/internal/app/services/user"
	"github.com/knstch/course/internal/app/storage"
)

type Handlers struct {
	authService  auth.AuthService
	userService  user.UserService
	emailService *email.EmailService
	address      string
}

func NewHandlers(storage *storage.Storage, config *config.Config, redisClient *redis.Client) *Handlers {
	emailService := email.NewEmailService(redisClient)
	return &Handlers{
		authService:  auth.NewAuthService(storage, config, redisClient, emailService),
		userService:  user.NewUserService(storage, emailService, redisClient),
		emailService: emailService,
		address:      config.Address,
	}
}

var (
	errBrokenJSON = errors.New("запрос передан в неверном формате")
)
