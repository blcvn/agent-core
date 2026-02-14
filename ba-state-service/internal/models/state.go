package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// WorkflowState represents possible states
type WorkflowState string

const (
	StatePending         WorkflowState = "PENDING"
	StateRunning         WorkflowState = "RUNNING"
	StateCompleted       WorkflowState = "COMPLETED"
	StateWaitingForInput WorkflowState = "WAITING_FOR_INPUT"
	StateResumed         WorkflowState = "RESUMED"
	StateFailed          WorkflowState = "FAILED"
	StateRetrying        WorkflowState = "RETRYING"
)

// StateData represents flexible state data
type StateData map[string]interface{}

func (s *StateData) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, s)
}

func (s StateData) Value() (driver.Value, error) {
	return json.Marshal(s)
}

// SessionState represents a session's state
type SessionState struct {
	ID        string        `gorm:"primaryKey" json:"id"`
	UserID    string        `gorm:"index" json:"user_id"`
	State     WorkflowState `json:"state"`
	Data      StateData     `gorm:"type:jsonb" json:"data"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

// WorkflowProgress tracks multi-step workflow
type WorkflowProgress struct {
	ID          uint          `gorm:"primaryKey" json:"id"`
	SessionID   string        `gorm:"index" json:"session_id"`
	WorkflowID  string        `json:"workflow_id"`
	CurrentStep int           `json:"current_step"`
	TotalSteps  int           `json:"total_steps"`
	State       WorkflowState `json:"state"`
	StepData    StateData     `gorm:"type:jsonb" json:"step_data"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

// StateSnapshot for checkpoint mechanism
type StateSnapshot struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	SessionID string    `gorm:"index" json:"session_id"`
	Snapshot  StateData `gorm:"type:jsonb" json:"snapshot"`
	Version   int       `json:"version"`
	CreatedAt time.Time `json:"created_at"`
}
