package storage

import (
	"errors"

	"github.com/knstch/course/internal/domain/dto"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	errRegistingUser = errors.New("ошибка при регистрации пользователя")
	errEmailIsBusy   = errors.New("пользователь с таким email уже существует")
	errUserNotFound  = errors.New("пользователь не найден")
	errTokenNotFound = errors.New("токен не найден")
)

type Storage struct {
	db     *gorm.DB
	secret string
}

func NewStorage(dsn, secret string) (*Storage, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return &Storage{
		db:     db,
		secret: secret,
	}, nil
}

func newUserProfileUpdate(firstName, surname string, phoneNumber int) map[string]interface{} {
	updates := make(map[string]interface{}, 3)

	updates["phone_number"] = phoneNumber
	updates["first_name"] = firstName
	updates["surname"] = surname

	return updates
}

func (storage *Storage) Automigrate() error {
	if err := storage.db.AutoMigrate(
		&dto.User{},
		&dto.Credentials{},
		&dto.Course{},
		&dto.AccessToken{},
	); err != nil {
		return err
	}

	return nil
}
