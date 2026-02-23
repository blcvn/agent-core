package main

import (
	"log"
	"os"

	"github.com/blcvn/backend/services/ba-interaction-service/internal/handlers"
	"github.com/blcvn/backend/services/ba-interaction-service/internal/ws"
	knowledgepb "github.com/blcvn/ba-shared-libs/proto/knowledge"
	orchestratorpb "github.com/blcvn/ba-shared-libs/proto/orchestrator"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// ── gRPC Client: Orchestrator ──
	orchestratorAddr := os.Getenv("ORCHESTRATOR_ADDR")
	if orchestratorAddr == "" {
		orchestratorAddr = "localhost:50052"
	}
	conn, err := grpc.Dial(orchestratorAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to Orchestrator Service: %v", err)
	}
	defer conn.Close()
	orchestratorClient := orchestratorpb.NewOrchestratorServiceClient(conn)

	// ── gRPC Client: Knowledge ──
	knowledgeAddr := os.Getenv("KNOWLEDGE_ADDR")
	if knowledgeAddr == "" {
		knowledgeAddr = "localhost:50053"
	}
	knowledgeConn, err := grpc.Dial(knowledgeAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to Knowledge Service: %v", err)
	}
	defer knowledgeConn.Close()
	knowledgeClient := knowledgepb.NewKnowledgeServiceClient(knowledgeConn)

	// ── gRPC Client: AI Agents ──
	agentsAddr := os.Getenv("AI_AGENTS_ADDR")
	if agentsAddr == "" {
		// Default port for ba-ai-agents-service. Orchestrator: 50052, Planner: 50051, Knowledge: 50053, Agents: 50054?
		// Let's assume 50054
		agentsAddr = "localhost:50054"
	}
	agentsConn, err := grpc.Dial(agentsAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to AI Agents Service: %v", err)
	}
	// connection is kept open
	defer agentsConn.Close()

	// ── Handlers ──
	agentHandler := handlers.NewAgentHandler(orchestratorClient)
	docHandler := handlers.NewDocumentHandler(knowledgeClient)

	// ── WebSocket Hub ──
	hub := ws.NewHub()
	go hub.Run()

	// ── Router ──
	r := gin.Default()

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "ba-interaction-service"})
	})

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Task endpoints
		v1.POST("/tasks", agentHandler.ExecuteTask)
		v1.GET("/tasks/:task_id", agentHandler.GetTaskStatus)
		v1.POST("/tasks/:task_id/input", agentHandler.SubmitInput)

		v1.POST("/tasks/:task_id/cancel", agentHandler.CancelTask)

		// Document endpoints
		v1.POST("/documents/prd", docHandler.CreatePRD)
		v1.POST("/documents/generate", docHandler.GenerateDocument)
		v1.GET("/documents/:id", docHandler.GetDocument)
		v1.PUT("/documents/:id", docHandler.UpdateDocument)
		v1.POST("/documents/:id/approve", docHandler.ApproveDocument)
		v1.POST("/documents/:id/review", docHandler.ReviewDocument)
		v1.POST("/documents/:id/regenerate", docHandler.RegenerateDocument)

		// Review endpoints
		v1.GET("/reviews/:id", docHandler.GetReviewStatus)
	}

	// WebSocket endpoint
	r.GET("/ws", func(c *gin.Context) {
		ws.ServeWs(hub, c.Writer, c.Request)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Interaction Service running on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
