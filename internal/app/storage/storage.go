package storage

import (
	"github.com/knstch/course/internal/domain/dto"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Storage struct {
	db *gorm.DB
}

func NewStorage(dsn string) (*Storage, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return &Storage{db: db}, nil
}

func (storage *Storage) Automigrate() error {
	if err := storage.db.AutoMigrate(
		&dto.User{},
		&dto.Credentials{},
		&dto.Subscription{},
		&dto.AccessToken{},
	); err != nil {
		return err
	}

	return nil
}
