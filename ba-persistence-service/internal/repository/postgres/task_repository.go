package postgres

import (
	"context"
	"encoding/json"
	"time"

	baagent "github.com/blcvn/kratos-proto/go/ba-agent"
	"gorm.io/gorm"
)

// AgentTaskModel represents the database schema for agent tasks
type AgentTaskModel struct {
	ID           string `gorm:"primaryKey;type:uuid"`
	Status       string `gorm:"index"`
	WorkflowMode string
	ResultData   []byte `gorm:"type:jsonb"` // Stores JSON representation of AgentTask (steps, response)
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (AgentTaskModel) TableName() string {
	return "agent_tasks"
}

// TaskRepository implements persistence for Agent Tasks
type TaskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

// CreateTask saves a new task to DB
func (r *TaskRepository) CreateTask(ctx context.Context, task *baagent.AgentTask, workflowMode string) error {
	data, err := json.Marshal(task)
	if err != nil {
		return err
	}

	model := &AgentTaskModel{
		ID:           task.Id,
		Status:       task.Status.String(),
		WorkflowMode: workflowMode,
		ResultData:   data,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	return r.db.WithContext(ctx).Create(model).Error
}

// UpdateTask updates an existing task
func (r *TaskRepository) UpdateTask(ctx context.Context, task *baagent.AgentTask) error {
	data, err := json.Marshal(task)
	if err != nil {
		return err
	}

	// We only update status, result_data, and updated_at
	return r.db.WithContext(ctx).Model(&AgentTaskModel{}).
		Where("id = ?", task.Id).
		Updates(map[string]interface{}{
			"status":      task.Status.String(),
			"result_data": data,
			"updated_at":  time.Now(),
		}).Error
}

// GetTask retrieves a task by ID
func (r *TaskRepository) GetTask(ctx context.Context, id string) (*baagent.AgentTask, string, error) {
	var model AgentTaskModel
	if err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		return nil, "", err
	}

	var task baagent.AgentTask
	if err := json.Unmarshal(model.ResultData, &task); err != nil {
		return nil, "", err
	}

	return &task, model.WorkflowMode, nil
}
