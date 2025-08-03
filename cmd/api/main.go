// @title PTOP API
// @version 1.0
// @description API сервиса PTOP
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"ptop/config"
	"ptop/internal/btcwatcher"
	"ptop/internal/db"
	"ptop/internal/ethwatcher"
	"ptop/internal/handlers"
	"ptop/internal/services"
	"ptop/internal/xmrwatcher"

	docs "ptop/docs"
)

func main() {
	// 1. Загружаем конфиг из .env / окружения
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config load failed: %v", err)
	}

	// 1.1 Определяем режим запуска (dev/prod)
	env := os.Getenv("APP_ENV")
	if env == "prod" || env == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// 2. Открываем GORM-подключение
	gormDB, err := db.NewDB(cfg.DSN)
	if err != nil {
		log.Fatalf("db connect failed: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr, Password: cfg.RedisPassword, DB: cfg.RedisDB})
	chatCache := services.NewChatCache(rdb, cfg.ChatCacheLimit)

	docs.SwaggerInfo.BasePath = "/"

	// 3. Создаём Gin-роутер и регистрируем /health
	r := gin.Default()
	r.GET("/health", handlers.Health(gormDB))
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	auth := r.Group("/auth")
	auth.POST("/register", handlers.Register(gormDB, cfg.TokenTypeTTL))
	auth.POST("/login", handlers.Login(gormDB, cfg.TokenTypeTTL))
	auth.POST("/refresh", handlers.Refresh(gormDB, cfg.TokenTypeTTL))
	auth.GET("/recover/:username", handlers.RecoverChallenge(gormDB))
	auth.POST("/recover", handlers.Recover(gormDB, cfg.TokenTypeTTL))
	auth.Use(handlers.AuthMiddleware(gormDB))
	auth.GET("/profile", handlers.Profile(gormDB))
	auth.POST("/username", handlers.ChangeUsername(gormDB))
	auth.POST("/pincode", handlers.SetPinCode(gormDB))
	auth.POST("/2fa/enable", handlers.Enable2FA(gormDB))
	auth.POST("/password", handlers.ChangePassword(gormDB))

	api := r.Group("/")
	api.Use(handlers.AuthMiddleware(gormDB))
	api.GET("/countries", handlers.GetCountries(gormDB))
	api.GET("/payment-methods", handlers.GetPaymentMethods(gormDB))
	api.GET("/assets", handlers.GetAssets(gormDB))
	api.GET("/client/payment-methods", handlers.ListClientPaymentMethods(gormDB))
	api.POST("/client/payment-methods", handlers.CreateClientPaymentMethod(gormDB))
	api.DELETE("/client/payment-methods/:id", handlers.DeleteClientPaymentMethod(gormDB))
	api.GET("/client/wallets", handlers.ListClientWallets(gormDB))
	api.POST("/client/wallets", handlers.CreateWallet(gormDB))

	api.GET("/offers", handlers.ListOffers(gormDB))
	api.GET("/client/offers", handlers.ListClientOffers(gormDB))
	api.POST("/client/offers", handlers.CreateOffer(gormDB))
	api.PUT("/client/offers/:id", handlers.UpdateOffer(gormDB))
	api.POST("/client/offers/:id/enable", handlers.EnableOffer(gormDB, cfg.MaxActiveOffersPerClient))
	api.POST("/client/offers/:id/disable", handlers.DisableOffer(gormDB))

	ws := r.Group("/ws")
	ws.Use(handlers.AuthMiddleware(gormDB))
	ws.GET("/orders/:id/chat", handlers.OrderChatWS(gormDB, chatCache))

	if cfg.WatchersDebug {
		btcW, err := btcwatcher.New(gormDB, cfg.BtcRPCHost, cfg.BtcRPCUser, cfg.BtcRPCPass, nil, true)
		if err != nil {
			log.Fatalf("btc watcher: %v", err)
		}
		if err := btcW.Start(); err != nil {
			log.Fatalf("btc watcher start: %v", err)
		}
		ethW, err := ethwatcher.New(gormDB, cfg.EthRPCURL, true)
		if err != nil {
			log.Fatalf("eth watcher: %v", err)
		}
		if err := ethW.Start(); err != nil {
			log.Fatalf("eth watcher start: %v", err)
		}
		xmrW, err := xmrwatcher.New(gormDB, cfg.MoneroRPCURL, 0, true)
		if err != nil {
			log.Fatalf("xmr watcher: %v", err)
		}
		xmrW.Start()
		watchers := map[string]handlers.DebugDepositor{
			"BTC": btcW,
			"ETH": ethW,
			"XMR": xmrW,
		}
		r.POST("/debug/deposit", handlers.DebugDeposit(gormDB, watchers))
	}

	// 4. Запускаем сервер
	addr := ":" + cfg.Port
	log.Printf("listening on %s …", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
