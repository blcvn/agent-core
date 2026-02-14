package handler

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/blcvn/backend/services/ba-orchestrator-service/internal/models"
	"github.com/blcvn/backend/services/pkg/queue"
	"gorm.io/gorm"
)

type Orchestrator struct {
	db       *gorm.DB
	producer *queue.Producer
}

func NewOrchestrator(db *gorm.DB, producer *queue.Producer) *Orchestrator {
	return &Orchestrator{
		db:       db,
		producer: producer,
	}
}

// CreateGoal creates a new high-level goal
func (o *Orchestrator) CreateGoal(goal *models.Goal) error {
	goal.Status = models.TaskStatusPending
	goal.CreatedAt = time.Now()
	return o.db.Create(goal).Error
}

// DecomposeGoal breaks down a goal into sub-goals and tasks
func (o *Orchestrator) DecomposeGoal(ctx context.Context, goalID string) ([]*models.Task, error) {
	var goal models.Goal
	if err := o.db.Where("id = ?", goalID).First(&goal).Error; err != nil {
		return nil, err
	}

	// Simple decomposition logic (in production, this would call Planner Service)
	tasks := make([]*models.Task, 0)

	for i, subGoal := range goal.SubGoals {
		task := &models.Task{
			ID:        fmt.Sprintf("%s-task-%d", goalID, i+1),
			SessionID: goalID,
			GoalID:    goalID,
			Type:      "sub_goal",
			Status:    models.TaskStatusPending,
			Input: models.JSONData{
				"description": subGoal,
			},
			Priority:   i,
			MaxRetries: 3,
			CreatedAt:  time.Now(),
		}
		tasks = append(tasks, task)

		if err := o.db.Create(task).Error; err != nil {
			return nil, err
		}
	}

	return tasks, nil
}

// SelectAgent selects the best agent for a task
func (o *Orchestrator) SelectAgent(ctx context.Context, task *models.Task) (string, error) {
	// Simple selection logic (in production, this would call Agent Registry + Optimizer)
	// For now, return a default agent based on task type

	agentMap := map[string]string{
		"index:prd":   "ba-expert-agent",
		"gen:outline": "document-generator-agent",
		"gen:code":    "coder-agent",
		"gen:diagram": "architect-agent",
		"sub_goal":    "ba-expert-agent",
	}

	agentID := agentMap[task.Type]
	if agentID == "" {
		agentID = "default-agent"
	}

	// Record selection
	selection := &models.AgentSelection{
		TaskID:     task.ID,
		AgentID:    agentID,
		Score:      0.85,
		Reason:     fmt.Sprintf("Selected based on task type: %s", task.Type),
		SelectedAt: time.Now(),
	}
	o.db.Create(selection)

	return agentID, nil
}

// ExecuteTask orchestrates task execution
func (o *Orchestrator) ExecuteTask(ctx context.Context, taskID string) error {
	var task models.Task
	if err := o.db.Where("id = ?", taskID).First(&task).Error; err != nil {
		return err
	}

	// Select agent
	agentID, err := o.SelectAgent(ctx, &task)
	if err != nil {
		return err
	}

	task.AgentID = agentID
	task.Status = models.TaskStatusRunning
	now := time.Now()
	task.StartedAt = &now
	o.db.Save(&task)

	// Enqueue task to appropriate queue
	payload := queue.GenCodePayload{
		TaskID:      task.ID,
		FeatureSpec: fmt.Sprintf("%v", task.Input["description"]),
		Language:    "go",
	}

	if err := o.producer.Enqueue(ctx, queue.TaskTypeGenCode, payload, 0); err != nil {
		task.Status = models.TaskStatusFailed
		task.Error = err.Error()
		o.db.Save(&task)
		return err
	}

	log.Printf("[Orchestrator] Task %s enqueued for agent %s", taskID, agentID)
	return nil
}

// CompleteTask marks a task as completed
func (o *Orchestrator) CompleteTask(taskID string, output models.JSONData) error {
	var task models.Task
	if err := o.db.Where("id = ?", taskID).First(&task).Error; err != nil {
		return err
	}

	task.Status = models.TaskStatusCompleted
	task.Output = output
	now := time.Now()
	task.CompletedAt = &now

	return o.db.Save(&task).Error
}

// FailTask marks a task as failed and handles retry logic
func (o *Orchestrator) FailTask(taskID string, errorMsg string) error {
	var task models.Task
	if err := o.db.Where("id = ?", taskID).First(&task).Error; err != nil {
		return err
	}

	task.RetryCount++
	task.Error = errorMsg

	if task.RetryCount >= task.MaxRetries {
		task.Status = models.TaskStatusFailed
	} else {
		task.Status = models.TaskStatusPending
		log.Printf("[Orchestrator] Task %s will retry (%d/%d)", taskID, task.RetryCount, task.MaxRetries)
	}

	return o.db.Save(&task).Error
}

// RequestHumanInput implements human-in-the-loop
func (o *Orchestrator) RequestHumanInput(taskID string, question string) (*models.HumanFeedback, error) {
	feedback := &models.HumanFeedback{
		TaskID:   taskID,
		Question: question,
		AskedAt:  time.Now(),
	}

	if err := o.db.Create(feedback).Error; err != nil {
		return nil, err
	}

	// Update task status
	o.db.Model(&models.Task{}).
		Where("id = ?", taskID).
		Update("status", models.TaskStatusWaiting)

	return feedback, nil
}

// ProvideHumanFeedback records human response
func (o *Orchestrator) ProvideHumanFeedback(feedbackID uint, response string) error {
	var feedback models.HumanFeedback
	if err := o.db.Where("id = ?", feedbackID).First(&feedback).Error; err != nil {
		return err
	}

	now := time.Now()
	feedback.Response = response
	feedback.RespondedAt = &now

	if err := o.db.Save(&feedback).Error; err != nil {
		return err
	}

	// Resume task
	o.db.Model(&models.Task{}).
		Where("id = ?", feedback.TaskID).
		Update("status", models.TaskStatusPending)

	return nil
}
