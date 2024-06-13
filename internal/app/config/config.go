package config

import (
	"log"
	"path/filepath"
	"strings"
	"sync"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Port                  string `envconfig:"PORT"`
	DSN                   string `envconfig:"DSN"`
	Secret                string `envconfig:"SECRET"`
	JwtSecret             string `envconfig:"JWT_SECRET"`
	RedisEmailChannelName string `envconfig:"REDIS_EMAIL_CHANNEL_NAME"`
	RedisDSN              string `envconfig:"REDIS_DSN"`
	Address               string `envconfig:"ADDRESS"`
	CdnHost               string `envconfig:"CDN_HOST"`
	CdnApiKey             string `envconfig:"CDN_API_KEY"`
}

var (
	config Config
	once   sync.Once
)

func GetConfig() *Config {
	once.Do(func() {
		err := envconfig.Process("", &config)
		if err != nil {
			log.Fatal(err)
		}
	})
	return &config
}

func InitENV(dir string) error {
	if err := godotenv.Load(filepath.Join(dir, ".env.local")); err != nil {
		if strings.Contains(err.Error(), "no such file or directory") {
			log.Print("файл .env.local не найден")
		} else {
			return err
		}
	}
	if err := godotenv.Load(filepath.Join(dir, ".env")); err != nil {
		return err
	}
	return nil
}
