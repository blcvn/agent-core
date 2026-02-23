package main

import (
	"log"
	"os"
	"strconv"

	"github.com/blcvn/backend/services/ba-orchestrator-service/internal/handler"
	"github.com/blcvn/backend/services/ba-orchestrator-service/internal/models"
	"github.com/blcvn/ba-shared-libs/pkg/queue"
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
		&models.Task{},
		&models.Goal{},
		&models.AgentSelection{},
		&models.HumanFeedback{},
	); err != nil {
		log.Fatalf("failed to migrate: %v", err)
	}

	// Initialize queue producer
	redisDB, _ := strconv.Atoi(os.Getenv("REDIS_DB"))
	queueCfg := queue.RedisConfig{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       redisDB,
	}
	producer := queue.NewProducer(queueCfg)

	orchestrator := handler.NewOrchestrator(db, producer)

	log.Println("Orchestrator Service started")
	_ = orchestrator

	// Keep service running
	select {}
}
