package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/blcvn/backend/services/ba-planner-service/internal/entities"
	"github.com/google/uuid"
)

type iRedisClient interface {
	SaveSession(ctx context.Context, session *entities.Session) error
	GetSession(ctx context.Context, id string) (*entities.Session, error)
	AddStep(ctx context.Context, sessionID string, step entities.Step) error
}

type iServiceBridge interface {
	RenderPrompt(ctx context.Context, name string, variables map[string]string) (string, error)
	Complete(ctx context.Context, modelID, systemPrompt, userPrompt string) (string, error)
}

type ToolFunc func(ctx context.Context, args map[string]interface{}) (interface{}, error)

type ReActEngine struct {
	redis  iRedisClient
	bridge iServiceBridge
	tools  map[string]ToolFunc
}

func NewReActEngine(redis iRedisClient, bridge iServiceBridge) *ReActEngine {
	return &ReActEngine{
		redis:  redis,
		bridge: bridge,
		tools:  make(map[string]ToolFunc),
	}
}

func (e *ReActEngine) RegisterTool(name string, fn ToolFunc) {
	e.tools[name] = fn
}

func (e *ReActEngine) ExecuteChat(ctx context.Context, sessionID, userInput string) (*entities.Step, error) {
	session, err := e.redis.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// 1. Add User Input to History
	userStep := entities.Step{
		ID:        uuid.New().String(),
		Type:      entities.StepObservation,
		Content:   userInput,
		Timestamp: time.Now(),
	}
	session.History = append(session.History, userStep)

	// 2. Think (Call LLM)
	// For Discovery Mode, we usually use a specific system prompt
	systemPrompt, err := e.bridge.RenderPrompt(ctx, "ba-discovery-system", nil)
	if err != nil {
		return nil, err
	}

	// Build user prompt from history
	// Simple version for now: just the last input
	llmResp, err := e.bridge.Complete(ctx, "claude-3-5-sonnet", systemPrompt, userInput)
	if err != nil {
		return nil, err
	}

	// 3. Create Assistant Step
	assistantStep := entities.Step{
		ID:        uuid.New().String(),
		Type:      entities.StepQuestion,
		Content:   llmResp,
		Timestamp: time.Now(),
	}
	session.History = append(session.History, assistantStep)
	session.UpdatedAt = time.Now()

	err = e.redis.SaveSession(ctx, session)
	return &assistantStep, err
}

func (e *ReActEngine) ExecutePipeline(ctx context.Context, sessionID string, workflowType string) error {
	// Pipeline mode implementation (Autonomous)
	// Will handle ANALYSIS, DOCUMENTATION phases
	// Loop: Think -> Act (Tool) -> Observe -> Repeat
	return fmt.Errorf("pipeline mode not implemented yet")
}
