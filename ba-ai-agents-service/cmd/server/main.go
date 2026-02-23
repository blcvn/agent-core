package main

import (
	"log"
	"net"
	"os"

	"github.com/blcvn/backend/services/ba-ai-agents-service/internal/handler"
	"github.com/blcvn/backend/services/ba-ai-agents-service/internal/models"
	"github.com/blcvn/backend/services/ba-ai-agents-service/internal/server"
	aiagentspb "github.com/blcvn/ba-shared-libs/proto/ai_agents"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
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

	// Initialize service logic
	svcLogic := handler.NewAIAgentsService(db)

	// Initialize gRPC server
	port := os.Getenv("PORT")
	if port == "" {
		port = "50053"
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	aiAgentsServer := server.NewAIAgentsServer(svcLogic)
	aiagentspb.RegisterAIAgentsServiceServer(s, aiAgentsServer)
	reflection.Register(s)

	log.Printf("AI Agents Service listening on :%s", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
