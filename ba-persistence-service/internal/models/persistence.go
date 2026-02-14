package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// JSONPayload represents flexible JSON data
type JSONPayload map[string]interface{}

func (j *JSONPayload) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}

func (j JSONPayload) Value() (driver.Value, error) {
	return json.Marshal(j)
}

// AgentInteraction captures input/output for an agent interaction
type AgentInteraction struct {
	ID          uint        `gorm:"primaryKey" json:"id"`
	SessionID   string      `gorm:"index" json:"session_id"`
	UserID      string      `gorm:"index" json:"user_id"`
	AgentID     string      `gorm:"index" json:"agent_id"`
	TaskID      string      `gorm:"index" json:"task_id"`
	Input       JSONPayload `gorm:"type:jsonb" json:"input"`
	Output      JSONPayload `gorm:"type:jsonb" json:"output"`
	InputSize   int         `json:"input_size"`  // bytes
	OutputSize  int         `json:"output_size"` // bytes
	Version     int         `json:"version"`
	StoragePath string      `json:"storage_path"` // S3/MinIO path for large payloads
	CreatedAt   time.Time   `json:"created_at"`
}

// PerformanceMetrics captures performance data
type PerformanceMetrics struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	InteractionID uint      `gorm:"index" json:"interaction_id"`
	LatencyMs     int       `json:"latency_ms"`
	TokensInput   int       `json:"tokens_input"`
	TokensOutput  int       `json:"tokens_output"`
	TokensTotal   int       `json:"tokens_total"`
	CostUSD       float64   `json:"cost_usd"`
	ModelUsed     string    `json:"model_used"`
	CreatedAt     time.Time `json:"created_at"`
}

// ErrorLog captures error information
type ErrorLog struct {
	ID            uint        `gorm:"primaryKey" json:"id"`
	InteractionID uint        `gorm:"index" json:"interaction_id"`
	ErrorType     string      `json:"error_type"`
	ErrorMessage  string      `gorm:"type:text" json:"error_message"`
	StackTrace    string      `gorm:"type:text" json:"stack_trace"`
	Context       JSONPayload `gorm:"type:jsonb" json:"context"`
	CreatedAt     time.Time   `json:"created_at"`
}

// AuditLog for compliance tracking
type AuditLog struct {
	ID         uint        `gorm:"primaryKey" json:"id"`
	UserID     string      `gorm:"index" json:"user_id"`
	Action     string      `json:"action"` // CREATE, READ, UPDATE, DELETE
	Resource   string      `json:"resource"`
	ResourceID string      `json:"resource_id"`
	Changes    JSONPayload `gorm:"type:jsonb" json:"changes"`
	IPAddress  string      `json:"ip_address"`
	UserAgent  string      `json:"user_agent"`
	CreatedAt  time.Time   `json:"created_at"`
}
