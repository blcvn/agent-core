package main

import (
	"log"
	"os"

	"github.com/blcvn/backend/services/ba-persistence-service/internal/handler"
	"github.com/blcvn/backend/services/ba-persistence-service/internal/models"
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
		&models.AgentInteraction{},
		&models.PerformanceMetrics{},
		&models.ErrorLog{},
		&models.AuditLog{},
	); err != nil {
		log.Fatalf("failed to migrate: %v", err)
	}

	service := handler.NewPersistenceService(db)

	log.Println("Data Persistence Service started")
	_ = service

	// Keep service running
	select {}
}
