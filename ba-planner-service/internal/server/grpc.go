package server

import (
	"context"
	"fmt"
	"log"

	engine "github.com/blcvn/backend/services/ba-planner-service/internal/engine/react"
	plannerpb "github.com/blcvn/backend/services/proto/planner"
	baagent "github.com/blcvn/kratos-proto/go/ba-agent"
)

type PlannerServer struct {
	plannerpb.UnimplementedPlannerServiceServer
	engine *engine.ReActEngine
}

func NewPlannerServer(reactEngine *engine.ReActEngine) *PlannerServer {
	return &PlannerServer{
		engine: reactEngine,
	}
}

func (s *PlannerServer) GeneratePlan(ctx context.Context, req *plannerpb.GeneratePlanRequest) (*plannerpb.GeneratePlanResponse, error) {
	// TODO: Implement planning logic (e.g. initial plan generation without execution)
	log.Printf("GeneratePlan called for session %s - STUB", req.SessionId)
	return &plannerpb.GeneratePlanResponse{
		PlanId:    "stub-plan-id",
		Reasoning: "Planning is currently integrated into ExecutePlan via ReAct loop.",
	}, nil
}

func (s *PlannerServer) ExecutePlan(req *plannerpb.ExecuteRequest, stream plannerpb.PlannerService_ExecutePlanServer) error {
	log.Printf("ExecutePlan called: %s", req.PlanId)

	// Extract task from context map or use a default
	task := "Analyze requirements and generate documentation" // Default
	if val, ok := req.Context["task"]; ok {
		task = val
	}

	// Map Request History to Engine History
	var history []*baagent.ReActStep
	for _, s := range req.History {
		history = append(history, &baagent.ReActStep{
			StepNumber:   s.StepNumber,
			Thought:      s.Thought,
			Action:       s.Action,
			ActionParams: map[string]string{"input": s.ActionInput},
			Observation:  s.Observation,
		})
	}

	payload := &baagent.ExecuteTaskPayload{
		SessionId: req.SessionId,
		History:   history,
	}

	// Callback for streaming
	onStep := func(step *engine.ReActStep) {
		status := "RUNNING"
		if step.IsFinal {
			status = "COMPLETED"
		}

		err := stream.Send(&plannerpb.ExecuteResponse{
			StepId: "step-generated", // We don't have ID in ReActStep yet
			Status: status,
			Output: fmt.Sprintf("Thought: %s\nAction: %s\nInput: %s\nObservation: %s",
				step.Thought, step.Action, step.ActionInput, step.Observation),
		})
		if err != nil {
			log.Printf("Failed to stream step: %v", err)
		}
	}

	result, err := s.engine.Execute(stream.Context(), task, payload, onStep)
	if err != nil {
		return fmt.Errorf("engine execution failed: %w", err)
	}

	if result.Status == "WAITING_FOR_INPUT" {
		// Send a final message indicating waiting for input
		if err := stream.Send(&plannerpb.ExecuteResponse{
			StepId: "waiting-input",
			Status: "WAITING_FOR_INPUT",
			Output: result.CurrentQuestion,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (s *PlannerServer) UpdatePlan(ctx context.Context, req *plannerpb.UpdatePlanRequest) (*plannerpb.UpdatePlanResponse, error) {
	return &plannerpb.UpdatePlanResponse{Success: true}, nil
}
