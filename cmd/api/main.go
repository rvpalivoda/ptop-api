// main.go
package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"ptop/config"
	"ptop/internal/db"
	"ptop/internal/handlers"
)

func main() {
	// 1. Загружаем конфиг из .env / окружения
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config load failed: %v", err)
	}

	// 2. Открываем GORM-подключение
	gormDB, err := db.NewDB(cfg.DSN)
	if err != nil {
		log.Fatalf("db connect failed: %v", err)
	}

	// 3. Создаём Gin-роутер и регистрируем /health
	r := gin.Default()
	r.GET("/health", handlers.Health(gormDB))

	// (здесь можно добавить другие роуты)

	// 4. Запускаем сервер
	addr := ":" + cfg.Port
	log.Printf("listening on %s …", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
