package main

import (
	"log"
	"net"
	"os"

	engine "github.com/blcvn/backend/services/ba-planner-service/internal/engine/react"
	"github.com/blcvn/backend/services/ba-planner-service/internal/server"
	executorpb "github.com/blcvn/backend/services/proto/executor"
	plannerpb "github.com/blcvn/backend/services/proto/planner"
	aiproxy "github.com/blcvn/kratos-proto/go/ai-proxy"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "50051"
	}

	aiProxyAddr := os.Getenv("AI_PROXY_ADDR")
	if aiProxyAddr == "" {
		aiProxyAddr = "localhost:50054"
	}

	executorAddr := os.Getenv("EXECUTOR_ADDR")
	if executorAddr == "" {
		executorAddr = "localhost:50053"
	}

	// Connect to AI Proxy Service
	aiProxyConn, err := grpc.Dial(aiProxyAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to AI Proxy Service: %v", err)
	}
	defer aiProxyConn.Close()
	aiProxyClient := aiproxy.NewAIProxyServiceClient(aiProxyConn)

	// Connect to Executor Service
	executorConn, err := grpc.Dial(executorAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to Executor Service: %v", err)
	}
	defer executorConn.Close()
	executorClient := executorpb.NewExecutorServiceClient(executorConn)

	// Initialize ReAct Engine
	reactEngine := engine.NewReActEngine(aiProxyClient, executorClient, "gpt-4", 10)

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	// Register service here
	plannerServer := server.NewPlannerServer(reactEngine)
	plannerpb.RegisterPlannerServiceServer(s, plannerServer)
	reflection.Register(s)

	log.Printf("Planner Service listening on :%s", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
