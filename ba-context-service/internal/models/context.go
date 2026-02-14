package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
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

// UserContext stores user-specific context
type UserContext struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      string    `gorm:"uniqueIndex;not null" json:"user_id"`
	Preferences JSONData  `gorm:"type:jsonb" json:"preferences"`
	History     JSONData  `gorm:"type:jsonb" json:"history"`
	Patterns    JSONData  `gorm:"type:jsonb" json:"patterns"` // behavioral patterns
	UpdatedAt   time.Time `json:"updated_at"`
}

// ProjectContext stores project-specific context
type ProjectContext struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	ProjectID   string    `gorm:"uniqueIndex;not null" json:"project_id"`
	Metadata    JSONData  `gorm:"type:jsonb" json:"metadata"`
	Goals       JSONData  `gorm:"type:jsonb" json:"goals"`
	Constraints JSONData  `gorm:"type:jsonb" json:"constraints"`
	Documents   []string  `gorm:"type:text[]" json:"documents"`
	Version     int       `json:"version"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// AggregatedContext combines user and project context
type AggregatedContext struct {
	UserContext    *UserContext    `json:"user_context"`
	ProjectContext *ProjectContext `json:"project_context"`
	Timestamp      time.Time       `json:"timestamp"`
}
