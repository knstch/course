package app

import (
	"net/http"

	"github.com/go-redis/redis"
	"github.com/knstch/course/internal/app/config"
	"github.com/knstch/course/internal/app/grpc"
	"github.com/knstch/course/internal/app/handlers"
	"github.com/knstch/course/internal/app/logger"
	authmiddleware "github.com/knstch/course/internal/app/middleware/auth_middleware"
	"github.com/knstch/course/internal/app/services/token"
	"github.com/knstch/course/internal/app/storage"
)

// Container используется для сборки проекта.
type Container struct {
	Storage    *storage.Storage
	Handlers   *handlers.Handlers
	Middleware *authmiddleware.Middleware
}

// InitContainer инициализирует контейнер, в качестве параметра принимает конфиг и возвращает готовый контейнер или ошибку.
func InitContainer(dir string, config *config.Config) (*Container, error) {
	psqlStorage, err := storage.NewStorage(config.DSN, config.Secret)
	if err != nil {
		return nil, err
	}

	if err := psqlStorage.Automigrate(config); err != nil {
		return nil, err
	}

	dsnRedis, err := redis.ParseURL(config.RedisDSN)
	if err != nil {
		return nil, err
	}

	redisClient := redis.NewClient(dsnRedis)

	httpClient := &http.Client{}

	grpcClient, err := grpc.NewGrpcClient(config)
	if err != nil {
		return nil, err
	}

	defaultLogger, err := logger.InitLogger(config.LogFileName)
	if err != nil {
		return nil, err
	}

	tokenService := token.NewTokenService(psqlStorage, config)

	middlware := authmiddleware.NewMiddleware(defaultLogger, config, tokenService)

	handlers := handlers.NewHandlers(psqlStorage, config, redisClient, httpClient, grpcClient, defaultLogger)

	return &Container{
		psqlStorage,
		handlers,
		middlware,
	}, nil
}
