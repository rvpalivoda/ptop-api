package config

import (
	"fmt"
	"os"
	"strconv"
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
	BtcRPCHost               string
	BtcRPCUser               string
	BtcRPCPass               string
	EthRPCURL                string
	MoneroRPCURL             string
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

	btcHost := os.Getenv("BTC_RPC_HOST")
	btcUser := os.Getenv("BTC_RPC_USER")
	btcPass := os.Getenv("BTC_RPC_PASS")
	ethURL := os.Getenv("ETH_RPC_URL")
	moneroURL := os.Getenv("MONERO_RPC_URL")

	return &Config{
		Port: port,
		DSN:  dsn,
		TokenTypeTTL: map[string]time.Duration{
			"access":  accessTTL,
			"refresh": refreshTTL,
		},
		MaxActiveOffersPerClient: maxOffers,
		WatchersDebug:            debug,
		BtcRPCHost:               btcHost,
		BtcRPCUser:               btcUser,
		BtcRPCPass:               btcPass,
		EthRPCURL:                ethURL,
		MoneroRPCURL:             moneroURL,
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
