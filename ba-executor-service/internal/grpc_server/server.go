package grpc_server

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/grpc"

	"github.com/blcvn/backend/services/ba-executor-service/internal/registry"
)

// ExecutorServer implements the grpc service
type ExecutorServer struct {
	regClient registry.Client
}

func RegisterExecutorServer(s *grpc.Server, reg registry.Client) {
	// TODO: Replace with actual RegisterExecutorServiceServer when proto is generated
	// pb.RegisterExecutorServiceServer(s, &ExecutorServer{regClient: reg})
	log.Println("Mock RegisterExecutorServer called")
}

func (s *ExecutorServer) ExecuteTool(ctx context.Context, toolName string, params map[string]interface{}) (string, error) {
	log.Printf("Received ExecuteTool request: %s", toolName)

	// 1. Lookup Tool Definition
	tools, err := s.regClient.ListTools()
	if err != nil {
		return "", fmt.Errorf("failed to list tools: %w", err)
	}

	found := false
	for _, t := range tools {
		if t.Name == toolName {
			found = true
			break
		}
	}

	if !found {
		return "", fmt.Errorf("tool %s not found in registry", toolName)
	}

	// 2. TODO: Implement actual execution logic (Local/MCP/HTTP)
	return fmt.Sprintf("Tool %s executed successfully (Mock)", toolName), nil
}
