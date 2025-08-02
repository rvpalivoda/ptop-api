package config

import (
        "fmt"
        "os"
        "time"

        "github.com/joho/godotenv"
)

// Config хранит все настройки приложения
type Config struct {
        Port        string
        DSN         string
        TokenTypeTTL map[string]time.Duration
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

        accessTTL := parseDuration(os.Getenv("TOKEN_TTL_ACCESS"), 15*time.Minute)
        refreshTTL := parseDuration(os.Getenv("TOKEN_TTL_REFRESH"), 7*24*time.Hour)

        return &Config{
                Port: port,
                DSN:  dsn,
                TokenTypeTTL: map[string]time.Duration{
                        "access":  accessTTL,
                        "refresh": refreshTTL,
                },
                // JWTSecret: os.Getenv("JWT_SECRET"),
                // Timezone:  os.Getenv("TIMEZONE"),
        }, nil
}

func parseDuration(val string, def time.Duration) time.Duration {
        if d, err := time.ParseDuration(val); err == nil {
                return d
        }
        return def
}
