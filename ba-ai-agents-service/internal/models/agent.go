package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// ReasoningStrategy defines how an agent reasons
type ReasoningStrategy string

const (
	StrategyReAct ReasoningStrategy = "ReAct"
	StrategyCoT   ReasoningStrategy = "CoT"
	StrategyToT   ReasoningStrategy = "ToT"
)

// JSONConfig represents flexible configuration
type JSONConfig map[string]interface{}

func (j *JSONConfig) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}

func (j JSONConfig) Value() (driver.Value, error) {
	return json.Marshal(j)
}

// AgentPersona represents an AI agent's persona and configuration
type AgentPersona struct {
	ID               string            `gorm:"primaryKey" json:"id"`
	Name             string            `gorm:"uniqueIndex;not null" json:"name"`
	Description      string            `json:"description"`
	Persona          string            `json:"persona"` // BA Expert, Architect, QA, etc.
	SystemPrompt     string            `gorm:"type:text" json:"system_prompt"`
	Strategy         ReasoningStrategy `json:"strategy"`
	Capabilities     []string          `gorm:"type:text[]" json:"capabilities"`
	Skills           []string          `gorm:"type:text[]" json:"skills"`
	KnowledgeDomains []string          `gorm:"type:text[]" json:"knowledge_domains"`
	BehaviorConfig   JSONConfig        `gorm:"type:jsonb" json:"behavior_config"`
	ModelConfig      JSONConfig        `gorm:"type:jsonb" json:"model_config"` // temperature, max_tokens, etc.
	IsActive         bool              `gorm:"default:true" json:"is_active"`
	Version          string            `json:"version"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
}

// Agent is an alias for AgentPersona
type Agent = AgentPersona

// AgentPerformance tracks agent quality metrics
type AgentPerformance struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	AgentID         string    `gorm:"index" json:"agent_id"`
	TasksCompleted  int       `json:"tasks_completed"`
	TasksSucceeded  int       `json:"tasks_succeeded"`
	TasksFailed     int       `json:"tasks_failed"`
	AvgQualityScore float64   `json:"avg_quality_score"` // 0-100
	AvgLatencyMs    int       `json:"avg_latency_ms"`
	TotalTokenUsage int64     `json:"total_token_usage"`
	TotalCostUSD    float64   `json:"total_cost_usd"`
	LastTaskAt      time.Time `json:"last_task_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// AgentSpecialization tracks domain-specific fine-tuning
type AgentSpecialization struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	AgentID   string     `gorm:"index" json:"agent_id"`
	Domain    string     `json:"domain"` // e.g., "healthcare", "finance"
	ModelPath string     `json:"model_path"`
	Config    JSONConfig `gorm:"type:jsonb" json:"config"`
	IsActive  bool       `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time  `json:"created_at"`
}
