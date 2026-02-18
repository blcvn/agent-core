package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/blcvn/backend/services/ba-orchestrator-service/internal/server"
	"github.com/blcvn/backend/services/ba-orchestrator-service/internal/usecases"
	"github.com/blcvn/backend/services/ba-orchestrator-service/internal/worker"
	"github.com/blcvn/backend/services/pkg/infrastructure/queue"
	orchestratorpb "github.com/blcvn/backend/services/proto/orchestrator"
	persistencepb "github.com/blcvn/backend/services/proto/persistence"
	plannerpb "github.com/blcvn/backend/services/proto/planner"
	"github.com/go-redis/redis/v8"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "50052"
	}

	// ── gRPC Client Connections ──
	persistenceAddr := os.Getenv("PERSISTENCE_ADDR")
	if persistenceAddr == "" {
		persistenceAddr = "localhost:50055"
	}
	persistenceConn, err := grpc.Dial(persistenceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to Persistence Service: %v", err)
	}
	defer persistenceConn.Close()
	persistenceClient := persistencepb.NewPersistenceServiceClient(persistenceConn)

	plannerAddr := os.Getenv("PLANNER_ADDR")
	if plannerAddr == "" {
		plannerAddr = "localhost:50051"
	}
	plannerConn, err := grpc.Dial(plannerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to Planner Service: %v", err)
	}
	defer plannerConn.Close()
	plannerClient := plannerpb.NewPlannerServiceClient(plannerConn)

	// ── Redis Queue ──
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	redisQueue := queue.NewRedisQueue(redisClient)

	// ── Usecase ──
	defaultModel := os.Getenv("DEFAULT_MODEL")
	if defaultModel == "" {
		defaultModel = "gpt-4"
	}
	agentUsecase := usecases.NewAgentUsecase(persistenceClient, redisQueue, defaultModel)

	// ── Worker ──
	agentWorker := worker.NewAgentWorker(redisQueue, persistenceClient, plannerClient, defaultModel)

	// Start worker in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	workerCount := 1 // Configurable
	for i := 0; i < workerCount; i++ {
		go agentWorker.Start(ctx)
	}
	log.Printf("Started %d agent worker(s)", workerCount)

	// ── gRPC Server ──
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	srv := server.NewOrchestratorServer(agentUsecase)
	orchestratorpb.RegisterOrchestratorServiceServer(s, srv)
	reflection.Register(s)

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("Shutting down Orchestrator Service...")
		cancel() // Stop workers
		s.GracefulStop()
	}()

	log.Printf("Orchestrator Service listening on :%s", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

	// Give workers time to finish current tasks
	time.Sleep(2 * time.Second)
}
