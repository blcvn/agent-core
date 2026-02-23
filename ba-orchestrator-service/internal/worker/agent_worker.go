package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/blcvn/ba-shared-libs/pkg/infrastructure/queue"
	persistencepb "github.com/blcvn/ba-shared-libs/proto/persistence"
	plannerpb "github.com/blcvn/ba-shared-libs/proto/planner"
	baagent "github.com/blcvn/kratos-proto/go/ba-agent"
)

type AgentWorker struct {
	queue             *queue.RedisQueue
	persistenceClient persistencepb.PersistenceServiceClient
	plannerClient     plannerpb.PlannerServiceClient
	defaultModel      string
}

func NewAgentWorker(
	queue *queue.RedisQueue,
	persistenceClient persistencepb.PersistenceServiceClient,
	plannerClient plannerpb.PlannerServiceClient,
	defaultModel string,
) *AgentWorker {
	return &AgentWorker{
		queue:             queue,
		persistenceClient: persistenceClient,
		plannerClient:     plannerClient,
		defaultModel:      defaultModel,
	}
}

// Start runs the worker loop
func (w *AgentWorker) Start(ctx context.Context) {
	log.Println("Starting Agent Worker...")
	for {
		select {
		case <-ctx.Done():
			log.Println("Agent Worker stopping...")
			return
		default:
			// Consume task (blocking)
			taskID, err := w.queue.Consume(ctx)
			if err != nil {
				log.Printf("Error consuming task: %v", err)
				time.Sleep(1 * time.Second) // Backoff
				continue
			}
			if taskID == "" {
				continue
			}

			// Process task synchronously to control concurrency
			w.processTask(ctx, taskID)
		}
	}
}

func (w *AgentWorker) processTask(ctx context.Context, taskID string) {
	// 1. Get Task from Persistence Service
	resp, err := w.persistenceClient.GetTask(ctx, &persistencepb.GetTaskRequest{TaskId: taskID})
	if err != nil {
		log.Printf("Failed to get task %s: %v", taskID, err)
		return
	}

	var task baagent.AgentTask
	if err := json.Unmarshal(resp.ResultDataJson, &task); err != nil {
		log.Printf("Failed to unmarshal task %s: %v", taskID, err)
		return
	}

	// 2. Update Status -> RUNNING
	task.Status = baagent.TaskStatus_TASK_RUNNING
	w.updateTaskInPersistence(ctx, taskID, resp.WorkflowMode, &task)

	// 3. Build history for Planner from existing steps
	var history []*plannerpb.ReActStep
	for _, s := range task.Steps {
		history = append(history, &plannerpb.ReActStep{
			StepNumber:  s.StepNumber,
			Thought:     s.Thought,
			Action:      s.Action,
			ActionInput: s.ActionParams["input"],
			Observation: s.Observation,
		})
	}

	// 4. Call Planner Service (streaming)
	stream, err := w.plannerClient.ExecutePlan(ctx, &plannerpb.ExecuteRequest{
		PlanId:    taskID,
		SessionId: task.SessionId,
		Context:   map[string]string{"task": task.TaskDescription},
		History:   history,
	})
	if err != nil {
		log.Printf("Failed to call Planner for task %s: %v", taskID, err)
		w.markTaskFailed(ctx, taskID, resp.WorkflowMode, &task, err)
		return
	}

	// 5. Handle Streaming Response
	var steps []*baagent.ReActStep
	stepNum := int32(len(task.Steps)) // Continue from existing steps
	finalStatus := baagent.TaskStatus_TASK_COMPLETED
	var finalOutput string

	for {
		planResp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Stream error for task %s: %v", taskID, err)
			w.markTaskFailed(ctx, taskID, resp.WorkflowMode, &task, err)
			return
		}

		switch planResp.Status {
		case "RUNNING":
			stepNum++
			step := &baagent.ReActStep{
				StepNumber:  stepNum,
				Observation: planResp.Output,
			}
			steps = append(steps, step)
			log.Printf("Task %s step %d: %s", taskID, stepNum, planResp.Output[:min(50, len(planResp.Output))])

		case "COMPLETED":
			finalOutput = planResp.Output
			finalStatus = baagent.TaskStatus_TASK_COMPLETED

		case "WAITING_FOR_INPUT":
			finalOutput = planResp.Output
			finalStatus = baagent.TaskStatus_TASK_WAITING_FOR_INPUT

		case "FAILED":
			log.Printf("Task %s failed from planner: %s", taskID, planResp.Error)
			w.markTaskFailed(ctx, taskID, resp.WorkflowMode, &task, fmt.Errorf(planResp.Error))
			return
		}
	}

	// 6. Update Task with results
	task.Steps = append(task.Steps, steps...)
	task.Status = finalStatus
	if finalStatus == baagent.TaskStatus_TASK_COMPLETED {
		task.FinalResponse = finalOutput
		task.IterationsUsed = stepNum
	}

	w.updateTaskInPersistence(ctx, taskID, resp.WorkflowMode, &task)
	log.Printf("Task %s finished with status: %s", taskID, finalStatus.String())
}

func (w *AgentWorker) updateTaskInPersistence(ctx context.Context, taskID, workflowMode string, task *baagent.AgentTask) {
	taskBytes, err := json.Marshal(task)
	if err != nil {
		log.Printf("Failed to marshal task %s: %v", taskID, err)
		return
	}

	_, err = w.persistenceClient.UpdateTask(ctx, &persistencepb.UpdateTaskRequest{
		TaskId:         taskID,
		Status:         task.Status.String(),
		ResultDataJson: taskBytes,
	})
	if err != nil {
		log.Printf("Failed to update task %s in Persistence: %v", taskID, err)
	}
}

func (w *AgentWorker) markTaskFailed(ctx context.Context, taskID, workflowMode string, task *baagent.AgentTask, failErr error) {
	task.Status = baagent.TaskStatus_TASK_FAILED
	task.FinalResponse = fmt.Sprintf("Task failed: %v", failErr)
	w.updateTaskInPersistence(ctx, taskID, workflowMode, task)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
