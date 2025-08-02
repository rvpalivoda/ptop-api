package main

import (
	"log"

	"ptop/config"
	"ptop/internal/db"
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

	if err := db.SeedCountries(gormDB); err != nil {
		log.Fatalf("seed countries failed: %v", err)
	}

	log.Println("countries seeded")
}
