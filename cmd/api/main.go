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
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"ptop/config"
	"ptop/internal/db"
	"ptop/internal/handlers"

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

	api.GET("/offers", handlers.ListOffers(gormDB))
	api.GET("/client/offers", handlers.ListClientOffers(gormDB))
	api.POST("/client/offers", handlers.CreateOffer(gormDB))
	api.PUT("/client/offers/:id", handlers.UpdateOffer(gormDB))
	api.POST("/client/offers/:id/enable", handlers.EnableOffer(gormDB, cfg.MaxActiveOffersPerClient))
	api.POST("/client/offers/:id/disable", handlers.DisableOffer(gormDB))

	// 4. Запускаем сервер
	addr := ":" + cfg.Port
	log.Printf("listening on %s …", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
