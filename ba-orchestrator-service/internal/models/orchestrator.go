package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// TaskStatus represents the status of a task
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusWaiting   TaskStatus = "waiting_for_input"
)

// JSONData represents flexible JSON data
type JSONData map[string]interface{}

func (j *JSONData) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}

func (j JSONData) Value() (driver.Value, error) {
	return json.Marshal(j)
}

// Task represents a unit of work in the orchestration
type Task struct {
	ID          string     `gorm:"primaryKey" json:"id"`
	SessionID   string     `gorm:"index" json:"session_id"`
	GoalID      string     `gorm:"index" json:"goal_id"`
	AgentID     string     `json:"agent_id"`
	Type        string     `json:"type"`
	Status      TaskStatus `json:"status"`
	Input       JSONData   `gorm:"type:jsonb" json:"input"`
	Output      JSONData   `gorm:"type:jsonb" json:"output"`
	Error       string     `json:"error,omitempty"`
	Priority    int        `gorm:"default:0" json:"priority"`
	RetryCount  int        `gorm:"default:0" json:"retry_count"`
	MaxRetries  int        `gorm:"default:3" json:"max_retries"`
	CreatedAt   time.Time  `json:"created_at"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// Goal represents a high-level objective
type Goal struct {
	ID          string     `gorm:"primaryKey" json:"id"`
	UserID      string     `gorm:"index" json:"user_id"`
	ProjectID   string     `gorm:"index" json:"project_id"`
	Description string     `gorm:"type:text" json:"description"`
	SubGoals    []string   `gorm:"type:text[]" json:"sub_goals"`
	Status      TaskStatus `json:"status"`
	Context     JSONData   `gorm:"type:jsonb" json:"context"`
	Result      JSONData   `gorm:"type:jsonb" json:"result"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// AgentSelection represents agent selection decision
type AgentSelection struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	TaskID     string    `gorm:"index" json:"task_id"`
	AgentID    string    `json:"agent_id"`
	Score      float64   `json:"score"` // Selection score
	Reason     string    `json:"reason"`
	SelectedAt time.Time `json:"selected_at"`
}

// HumanFeedback represents human-in-the-loop interactions
type HumanFeedback struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	TaskID      string     `gorm:"index" json:"task_id"`
	Question    string     `gorm:"type:text" json:"question"`
	Response    string     `gorm:"type:text" json:"response"`
	AskedAt     time.Time  `json:"asked_at"`
	RespondedAt *time.Time `json:"responded_at,omitempty"`
}
