package server

import (
	"context"
	"log"

	"github.com/blcvn/backend/services/ba-orchestrator-service/internal/usecases"
	orchestratorpb "github.com/blcvn/ba-shared-libs/proto/orchestrator"
)

type OrchestratorServer struct {
	orchestratorpb.UnimplementedOrchestratorServiceServer
	usecase *usecases.AgentUsecase
}

func NewOrchestratorServer(usecase *usecases.AgentUsecase) *OrchestratorServer {
	return &OrchestratorServer{
		usecase: usecase,
	}
}

func (s *OrchestratorServer) ExecuteTask(ctx context.Context, req *orchestratorpb.ExecuteTaskRequest) (*orchestratorpb.ExecuteTaskResponse, error) {
	log.Printf("ExecuteTask called for session %s (user %s)", req.SessionId, req.UserId)

	contextMap := map[string]string{
		"session_id": req.SessionId,
		"user_id":    req.UserId,
	}
	// Merge additional metadata
	for k, v := range req.Metadata {
		contextMap[k] = v
	}

	task, err := s.usecase.ExecuteTask(ctx, req.Content, contextMap, 15, "interactive")
	if err != nil {
		return nil, err
	}

	return &orchestratorpb.ExecuteTaskResponse{
		TaskId: task.Id,
		Status: task.Status.String(),
	}, nil
}

func (s *OrchestratorServer) GetTaskStatus(ctx context.Context, req *orchestratorpb.GetTaskStatusRequest) (*orchestratorpb.GetTaskStatusResponse, error) {
	task, err := s.usecase.GetTaskStatus(ctx, req.TaskId)
	if err != nil {
		return nil, err
	}

	return &orchestratorpb.GetTaskStatusResponse{
		TaskId: task.Id,
		Status: task.Status.String(),
		Result: task.FinalResponse,
	}, nil
}

func (s *OrchestratorServer) CancelTask(ctx context.Context, req *orchestratorpb.CancelTaskRequest) (*orchestratorpb.CancelTaskResponse, error) {
	// TODO: Implement task cancellation via Persistence + queue removal
	log.Printf("CancelTask called for task %s - TODO", req.TaskId)
	return &orchestratorpb.CancelTaskResponse{Success: true, Message: "Cancellation requested"}, nil
}

func (s *OrchestratorServer) SubmitInput(ctx context.Context, req *orchestratorpb.SubmitInputRequest) (*orchestratorpb.SubmitInputResponse, error) {
	_, err := s.usecase.SubmitInput(ctx, req.TaskId, req.InputData)
	if err != nil {
		return &orchestratorpb.SubmitInputResponse{Success: false, Message: err.Error()}, nil
	}
	return &orchestratorpb.SubmitInputResponse{Success: true, Message: "Input submitted, task resumed"}, nil
}
