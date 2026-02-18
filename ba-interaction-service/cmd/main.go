package main

import (
	"log"
	"os"

	"github.com/blcvn/backend/services/ba-interaction-service/internal/handlers"
	"github.com/blcvn/backend/services/ba-interaction-service/internal/ws"
	orchestratorpb "github.com/blcvn/backend/services/proto/orchestrator"
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

	// ── Handlers ──
	agentHandler := handlers.NewAgentHandler(orchestratorClient)

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
