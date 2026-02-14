package main

import (
	"log"
	"os"

	"github.com/blcvn/backend/services/ba-state-service/internal/handler"
	"github.com/blcvn/backend/services/ba-state-service/internal/models"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=ba_agent port=5432 sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// Auto-migrate schemas
	if err := db.AutoMigrate(
		&models.SessionState{},
		&models.WorkflowProgress{},
		&models.StateSnapshot{},
	); err != nil {
		log.Fatalf("failed to migrate: %v", err)
	}

	// Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})

	manager := handler.NewStateManager(db, rdb)

	log.Println("State Management Service started")
	_ = manager

	// Keep service running
	select {}
}
