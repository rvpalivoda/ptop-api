package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config хранит все настройки приложения
type Config struct {
    Port                     string
    DSN                      string
    TokenTypeTTL             map[string]time.Duration
    MaxActiveOffersPerClient int
    WatchersDebug            bool
    CORSAllowedOrigins       []string
    OrderExpirerInterval     time.Duration
	BtcRPCHost               string
	BtcRPCUser               string
	BtcRPCPass               string
	EthRPCURL                string
	MoneroRPCURL             string
	RedisAddr                string
	RedisPassword            string
	RedisDB                  int
	ChatCacheLimit           int64
	S3Endpoint               string
	S3AccessKey              string
	S3SecretKey              string
	S3Bucket                 string
	S3Region                 string
	S3UseSSL                 bool
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

	maxOffers := 3
	if v, err := strconv.Atoi(os.Getenv("MAX_ACTIVE_OFFERS")); err == nil {
		maxOffers = v
	}

	debug := os.Getenv("WATCHERS_DEBUG") == "1"

	corsEnv := os.Getenv("CORS_ALLOWED_ORIGINS")
	if corsEnv == "" {
		corsEnv = "http://localhost:5173"
	}
	corsOrigins := strings.Split(corsEnv, ",")

	btcHost := os.Getenv("BTC_RPC_HOST")
	btcUser := os.Getenv("BTC_RPC_USER")
	btcPass := os.Getenv("BTC_RPC_PASS")
	ethURL := os.Getenv("ETH_RPC_URL")
	moneroURL := os.Getenv("MONERO_RPC_URL")

	s3Endpoint := os.Getenv("S3_ENDPOINT")
	s3Access := os.Getenv("S3_ACCESS_KEY")
	s3Secret := os.Getenv("S3_SECRET_KEY")
	s3Bucket := os.Getenv("S3_BUCKET")
	s3Region := os.Getenv("S3_REGION")
	s3UseSSL := os.Getenv("S3_USE_SSL") == "1" || strings.ToLower(os.Getenv("S3_USE_SSL")) == "true"

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	redisPass := os.Getenv("REDIS_PASSWORD")
	redisDB := 0
	if v, err := strconv.Atoi(os.Getenv("REDIS_DB")); err == nil {
		redisDB = v
	}

    chatLimit := int64(50)
    if v, err := strconv.ParseInt(os.Getenv("CHAT_CACHE_LIMIT"), 10, 64); err == nil {
        chatLimit = v
    }

    // Интервал фоновой задачи авто-отмены ордеров
    expirerInterval := parseDuration(os.Getenv("ORDER_EXPIRER_INTERVAL"), 30*time.Second)

	return &Config{
		Port: port,
		DSN:  dsn,
		TokenTypeTTL: map[string]time.Duration{
			"access":  accessTTL,
			"refresh": refreshTTL,
		},
		MaxActiveOffersPerClient: maxOffers,
		WatchersDebug:            debug,
		CORSAllowedOrigins:       corsOrigins,
		BtcRPCHost:               btcHost,
		BtcRPCUser:               btcUser,
		BtcRPCPass:               btcPass,
		EthRPCURL:                ethURL,
		MoneroRPCURL:             moneroURL,
		RedisAddr:                redisAddr,
		RedisPassword:            redisPass,
		RedisDB:                  redisDB,
		ChatCacheLimit:           chatLimit,
		S3Endpoint:               s3Endpoint,
		S3AccessKey:              s3Access,
		S3SecretKey:              s3Secret,
		S3Bucket:                 s3Bucket,
        S3Region:                 s3Region,
        S3UseSSL:                 s3UseSSL,
        OrderExpirerInterval:     expirerInterval,
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
