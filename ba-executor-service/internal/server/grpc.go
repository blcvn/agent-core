package server

import (
	"context"
	"fmt"
	"time"

	"github.com/blcvn/backend/services/ba-executor-service/internal/tools"
	executorpb "github.com/blcvn/ba-shared-libs/proto/executor"
)

// ExecutorServer implements the gRPC server for tool execution
type ExecutorServer struct {
	executorpb.UnimplementedExecutorServiceServer
	toolManager *tools.ToolManager
}

// NewExecutorServer creates a new instance of ExecutorServer
func NewExecutorServer(tm *tools.ToolManager) *ExecutorServer {
	return &ExecutorServer{
		toolManager: tm,
	}
}

// ExecuteTool finds and executes a registered tool
func (s *ExecutorServer) ExecuteTool(ctx context.Context, req *executorpb.ExecuteToolRequest) (*executorpb.ExecuteToolResponse, error) {
	if req.ToolName == "" {
		return &executorpb.ExecuteToolResponse{
			Success: false,
			Error:   "tool name is required",
		}, nil
	}

	start := time.Now()

	// Execute the tool using the manager
	output, err := s.toolManager.Execute(ctx, req.ToolName, req.InputJson)

	duration := time.Since(start).Milliseconds()

	if err != nil {
		return &executorpb.ExecuteToolResponse{
			Success:    false,
			Error:      err.Error(),
			DurationMs: duration,
		}, nil
	}

	return &executorpb.ExecuteToolResponse{
		Success:    true,
		Output:     output,
		DurationMs: duration,
	}, nil
}

// ListTools returns a list of all available tools
func (s *ExecutorServer) ListTools(ctx context.Context, req *executorpb.ListToolsRequest) (*executorpb.ListToolsResponse, error) {
	toolsList := s.toolManager.List()
	var result []*executorpb.ToolInfo

	for _, t := range toolsList {
		result = append(result, &executorpb.ToolInfo{
			Name:        t.Name(),
			Description: t.Description(),
			// Category: t.Category(), // Not yet in interface
		})
	}

	return &executorpb.ListToolsResponse{
		Tools: result,
	}, nil
}

// GetToolSchema retrieves the JSON schema for a specific tool
func (s *ExecutorServer) GetToolSchema(ctx context.Context, req *executorpb.GetToolSchemaRequest) (*executorpb.GetToolSchemaResponse, error) {
	tool, exists := s.toolManager.Get(req.ToolName)
	if !exists {
		return nil, fmt.Errorf("tool not found: %s", req.ToolName)
	}

	return &executorpb.GetToolSchemaResponse{
		SchemaJson: tool.Schema(),
	}, nil
}
