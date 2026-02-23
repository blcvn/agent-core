package usecases

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/blcvn/ba-shared-libs/pkg/infrastructure/queue"
	persistencepb "github.com/blcvn/ba-shared-libs/proto/persistence"
	baagent "github.com/blcvn/kratos-proto/go/ba-agent"
	"github.com/google/uuid"
)

// AgentUsecase handles BA agent business logic
type AgentUsecase struct {
	persistenceClient persistencepb.PersistenceServiceClient
	queue             *queue.RedisQueue
	defaultModelID    string
}

// NewAgentUsecase creates a new agent usecase
func NewAgentUsecase(
	persistenceClient persistencepb.PersistenceServiceClient,
	queue *queue.RedisQueue,
	defaultModelID string,
) *AgentUsecase {
	return &AgentUsecase{
		persistenceClient: persistenceClient,
		queue:             queue,
		defaultModelID:    defaultModelID,
	}
}

// ExecuteTask executes a BA task using the appropriate engine based on mode
// In Async mode, this queues the task and returns immediately.
func (u *AgentUsecase) ExecuteTask(
	ctx context.Context,
	taskDescription string,
	contextMap map[string]string,
	maxIterations int,
	workflowMode string,
) (*baagent.AgentTask, error) {
	taskID := uuid.New().String()

	// 1. Initial State
	initialTask := &baagent.AgentTask{
		Id:              taskID,
		TaskDescription: taskDescription,
		Status:          baagent.TaskStatus_TASK_PENDING,
	}
	if sid, ok := contextMap["session_id"]; ok {
		initialTask.SessionId = sid
	}
	if uid, ok := contextMap["user_id"]; ok {
		initialTask.UserId = uid
	}

	// Serialize task to JSON
	taskBytes, err := json.Marshal(initialTask)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal task: %w", err)
	}

	// Save to Persistence Service
	req := &persistencepb.CreateTaskRequest{
		TaskId:         taskID,
		Status:         baagent.TaskStatus_TASK_PENDING.String(),
		WorkflowMode:   workflowMode,
		ResultDataJson: taskBytes,
	}

	_, err = u.persistenceClient.CreateTask(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create task in Persistence Service: %w", err)
	}

	// 2. Publish to Queue
	if err := u.queue.Publish(ctx, taskID); err != nil {
		return nil, fmt.Errorf("failed to publish task to queue: %w", err)
	}

	return initialTask, nil
}

// GetTaskStatus retrieves task status from Persistence Service
func (u *AgentUsecase) GetTaskStatus(ctx context.Context, taskID string) (*baagent.AgentTask, error) {
	resp, err := u.persistenceClient.GetTask(ctx, &persistencepb.GetTaskRequest{
		TaskId: taskID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get task from Persistence Service: %w", err)
	}

	var task baagent.AgentTask
	if err := json.Unmarshal(resp.ResultDataJson, &task); err != nil {
		return nil, fmt.Errorf("failed to unmarshal task data: %w", err)
	}

	// Ensure mapped status is correct if different from internal JSON
	// The tasks' internal JSON status should be the source of truth

	return &task, nil
}

// SubmitInput handles user input for a paused task
func (u *AgentUsecase) SubmitInput(ctx context.Context, taskID string, inputData string) (*baagent.AgentTask, error) {
	// 1. Retrieve task state
	resp, err := u.persistenceClient.GetTask(ctx, &persistencepb.GetTaskRequest{TaskId: taskID})
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	var task baagent.AgentTask
	if err := json.Unmarshal(resp.ResultDataJson, &task); err != nil {
		return nil, fmt.Errorf("failed to unmarshal task data: %w", err)
	}

	// 2. Validate status
	if task.Status != baagent.TaskStatus_TASK_WAITING_FOR_INPUT {
		// Try to parse string status if enum mismatch?
		// Assuming strict enum check is fine
		return nil, fmt.Errorf("task is not waiting for input (status: %v)", task.Status)
	}

	// 3. Update task with input (Observation)
	if len(task.Steps) > 0 {
		task.Steps[len(task.Steps)-1].Observation = inputData
	}

	// Reset status to PENDING
	task.Status = baagent.TaskStatus_TASK_PENDING

	// Re-serialize
	taskBytes, err := json.Marshal(&task)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal updated task: %w", err)
	}

	// Update Persistence
	_, err = u.persistenceClient.UpdateTask(ctx, &persistencepb.UpdateTaskRequest{
		TaskId:         taskID,
		Status:         baagent.TaskStatus_TASK_PENDING.String(),
		ResultDataJson: taskBytes,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	// 4. Re-Publish to Queue
	if err := u.queue.Publish(ctx, taskID); err != nil {
		return nil, fmt.Errorf("failed to re-publish task: %w", err)
	}

	return &task, nil
}
