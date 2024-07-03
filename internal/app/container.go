package app

import (
	"net/http"

	"github.com/go-redis/redis"
	"github.com/knstch/course/internal/app/config"
	"github.com/knstch/course/internal/app/grpc"
	"github.com/knstch/course/internal/app/handlers"
	"github.com/knstch/course/internal/app/logger"
	"github.com/knstch/course/internal/app/storage"
)

type Container struct {
	Storage  *storage.Storage
	Handlers *handlers.Handlers
}

func InitContainer(config *config.Config) (*Container, error) {
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

	handlers := handlers.NewHandlers(psqlStorage, config, redisClient, httpClient, grpcClient, defaultLogger)

	return &Container{
		Storage:  psqlStorage,
		Handlers: handlers,
	}, nil
}
