package main

import (
	"log"
	"os"

	"github.com/blcvn/backend/services/ba-context-service/internal/handler"
	"github.com/blcvn/backend/services/ba-context-service/internal/models"
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
	if err := db.AutoMigrate(&models.UserContext{}, &models.ProjectContext{}); err != nil {
		log.Fatalf("failed to migrate: %v", err)
	}

	// Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})

	service := handler.NewContextService(db, rdb)

	log.Println("Context Service started")
	_ = service

	// Keep service running
	select {}
}
