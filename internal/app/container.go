package app

import (
	"github.com/knstch/course/internal/app/config"
	"github.com/knstch/course/internal/app/storage"
)

type Container struct {
	Storage *storage.Storage
}

func InitContainer(config *config.Config) (*Container, error) {
	psqlStorage, err := storage.NewStorage(config.DSN)
	if err != nil {
		return nil, err
	}

	if err := psqlStorage.Automigrate(); err != nil {
		return nil, err
	}

	return &Container{
		Storage: psqlStorage,
	}, nil
}
