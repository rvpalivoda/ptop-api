package main

import (
	"log"

	"ptop/config"
	"ptop/internal/db"
	"ptop/internal/models"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config load failed: %v", err)
	}

	gormDB, err := db.NewDB(cfg.DSN)
	if err != nil {
		log.Fatalf("db connect failed: %v", err)
	}

	if err := gormDB.AutoMigrate(&models.Notification{}); err != nil {
		log.Fatalf("auto migrate failed: %v", err)
	}

	migrator := gormDB.Migrator()
	if err := migrator.CreateIndex(&models.Notification{}, "ClientID"); err != nil {
		log.Fatalf("create index failed: %v", err)
	}
	if err := migrator.CreateIndex(&models.Notification{}, "SentAt"); err != nil {
		log.Fatalf("create index failed: %v", err)
	}
	if err := migrator.CreateIndex(&models.Notification{}, "ReadAt"); err != nil {
		log.Fatalf("create index failed: %v", err)
	}

	log.Println("migration completed")
}
