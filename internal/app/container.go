package app

import (
	"github.com/go-redis/redis"
	"github.com/knstch/course/internal/app/config"
	"github.com/knstch/course/internal/app/handlers"
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

	if err := psqlStorage.Automigrate(); err != nil {
		return nil, err
	}

	dsnRedis, err := redis.ParseURL(config.RedisDSN)
	if err != nil {
		return nil, err
	}

	redisClient := redis.NewClient(dsnRedis)

	handlers := handlers.NewHandlers(psqlStorage, config, redisClient)

	return &Container{
		Storage:  psqlStorage,
		Handlers: handlers,
	}, nil
}
