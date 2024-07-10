package app

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/go-redis/redis"
	"github.com/knstch/course/internal/app/config"
	"github.com/knstch/course/internal/app/grpc"
	"github.com/knstch/course/internal/app/handlers"
	"github.com/knstch/course/internal/app/logger"
	"github.com/knstch/course/internal/app/storage"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// Container используется для сборки проекта.
type Container struct {
	Storage  *storage.Storage
	Handlers *handlers.Handlers
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
	file, err := os.ReadFile(fmt.Sprintf("%v%v", dir, config.CredentialsKeyPath))
	if err != nil {
		fmt.Println("EEEEEEE")
	}
	fmt.Println("FILE: ", file)

	creds, err := google.CredentialsFromJSON(context.TODO(), []byte(config.CredentialsKeyPath), gmail.MailGoogleComScope)
	if err != nil {
		fmt.Println("AAAAAAAAAAAA")
		return nil, err
	}

	gmailService, err := gmail.NewService(context.TODO(), option.WithCredentials(creds))
	if err != nil {
		fmt.Println("SSSSSSSSSSSSS")
		return nil, err
	}

	handlers := handlers.NewHandlers(psqlStorage, config, redisClient, httpClient, grpcClient, defaultLogger, gmailService)

	return &Container{
		Storage:  psqlStorage,
		Handlers: handlers,
	}, nil
}
