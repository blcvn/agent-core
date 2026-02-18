package main

import (
	"fmt"
	"log"
	"net"

	"github.com/blcvn/backend/services/ba-executor-service/internal/server"
	"github.com/blcvn/backend/services/ba-executor-service/internal/tools"
	"github.com/blcvn/backend/services/ba-executor-service/internal/tools/interaction"
	executorpb "github.com/blcvn/backend/services/proto/executor"

	aiproxy "github.com/blcvn/kratos-proto/go/ai-proxy"
	prompt "github.com/blcvn/kratos-proto/go/prompt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// 1. Initialize dependencies (TODO: Connect to AI Proxy and Prompt Service)
	// For now, we'll placeholder nil clients or mock clients if needed
	var aiProxyClient aiproxy.AIProxyServiceClient
	var promptClient prompt.PromptServiceClient
	_ = aiProxyClient
	_ = promptClient

	// 2. Initialize Tool Manager
	toolManager := tools.NewToolManager()

	// 3. Register Tools
	// Note: We need to instantiate tools with their dependencies
	// Assuming constructors like NewAnalysisTool(aiProxyClient, promptClient) exist
	// Since dependencies are placeholders, we might need to adjust tools or pass nil for now
	// This is a migration step - wiring up is crucial.

	// Example registration (uncomment once dependencies are resolved):
	// toolManager.Register(analysis.NewAnalysisTool(aiProxyClient, promptClient))
	// toolManager.Register(analyzer.NewFeatureAnalyzerTool(aiProxyClient))
	// toolManager.Register(interaction.NewAskUserTool())
	// toolManager.Register(confluence.NewConfluenceSearchTool())
	// toolManager.Register(diagram.NewDiagramGeneratorTool(aiProxyClient))
	// toolManager.Register(discovery.NewDiscoveryTool(aiProxyClient, promptClient))
	// toolManager.Register(document.NewDocumentGeneratorTool(aiProxyClient, promptClient))

	// For now, register interaction tool since it has no dependencies (or minimal)
	toolManager.Register(interaction.NewAskUserTool())

	// 4. Initialize gRPC Server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 50051)) // Default port
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	executorServer := server.NewExecutorServer(toolManager)
	executorpb.RegisterExecutorServiceServer(s, executorServer)

	// Enable reflection for debugging
	reflection.Register(s)

	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
