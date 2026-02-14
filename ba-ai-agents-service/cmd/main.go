package main

import (
	"log"
	"os"

	"github.com/blcvn/backend/services/ba-ai-agents-service/internal/handler"
	"github.com/blcvn/backend/services/ba-ai-agents-service/internal/models"
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
		&models.AgentPersona{},
		&models.AgentPerformance{},
		&models.AgentSpecialization{},
	); err != nil {
		log.Fatalf("failed to migrate: %v", err)
	}

	service := handler.NewAIAgentsService(db)

	log.Println("AI Agents Service started")
	log.Printf("Service initialized with %d personas", countPersonas(service))

	// Keep service running
	select {}
}

func countPersonas(s *handler.AIAgentsService) int {
	personas, _ := s.ListPersonas()
	return len(personas)
}
