package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config хранит все настройки приложения
type Config struct {
	Port string
	DSN  string
	// Другие поля, например:
	// JWTSecret string
	// Timezone  string
}

// Load читает .env (если есть) и возвращает заполненный Config
func Load() (*Config, error) {
	// Попробуем загрузить файл .env — если его нет, просто пропускаем
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		return nil, fmt.Errorf("DB_DSN must be set")
	}

	return &Config{
		Port: port,
		DSN:  dsn,
		// JWTSecret: os.Getenv("JWT_SECRET"),
		// Timezone:  os.Getenv("TIMEZONE"),
	}, nil
}
