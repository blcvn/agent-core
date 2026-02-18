package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/blcvn/backend/services/ba-persistence-service/internal/repository/approval"
	"github.com/blcvn/backend/services/ba-persistence-service/internal/repository/document"
	"github.com/blcvn/backend/services/ba-persistence-service/internal/repository/graph"
	"github.com/blcvn/backend/services/ba-persistence-service/internal/repository/review"
	"github.com/blcvn/backend/services/ba-persistence-service/internal/repository/task"
	"github.com/blcvn/backend/services/ba-persistence-service/internal/server"
	"github.com/blcvn/backend/services/pkg/infrastructure/postgres"
	persistencepb "github.com/blcvn/backend/services/proto/persistence"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// 1. Initialize Database
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	// Default values for local dev
	if dbHost == "" {
		dbHost = "localhost"
	}
	if dbPort == "" {
		dbPort = "5432"
	}
	if dbUser == "" {
		dbUser = "postgres"
	}
	if dbPass == "" {
		dbPass = "password"
	}
	if dbName == "" {
		dbName = "ba_agent"
	}

	pgDB, err := postgres.NewPostgresDB(dbHost, dbPort, dbUser, dbPass, dbName)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pgDB.Close()

	// 2. Initialize Repositories
	taskRepo := task.NewTaskRepository(pgDB.DB)
	docRepo := document.NewDocumentRepository(pgDB.DB)
	graphRepo := graph.NewGraphRepository(pgDB.DB)
	approvalRepo := approval.NewApprovalRepository(pgDB.DB)
	reviewRepo := review.NewReviewRepository(pgDB.DB)

	// 3. Initialize gRPC Server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 50052)) // Port 50052 for Persistence
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	persistenceServer := server.NewPersistenceServer(
		taskRepo,
		docRepo,
		approvalRepo,
		graphRepo,
		reviewRepo,
	)
	persistencepb.RegisterPersistenceServiceServer(s, persistenceServer)

	// Enable reflection
	reflection.Register(s)

	log.Printf("Persistence Service listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
