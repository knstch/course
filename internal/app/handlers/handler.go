package handlers

import (
	"errors"

	"github.com/go-redis/redis"
	"github.com/knstch/course/internal/app/config"
	"github.com/knstch/course/internal/app/services/auth"
	"github.com/knstch/course/internal/app/storage"
)

type Handlers struct {
	authService auth.AuthService
	address     string
}

func NewHandlers(storage *storage.Storage, config *config.Config, redisClient *redis.Client) *Handlers {
	return &Handlers{
		authService: auth.NewAuthService(storage, config, redisClient),
		address:     config.Address,
	}
}

var (
	errBrokenJSON = errors.New("запрос передан в неверном формате")
)
