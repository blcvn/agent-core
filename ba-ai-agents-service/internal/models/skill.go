package models

import (
	"time"
)

type Skill struct {
	ID          string     `gorm:"primaryKey" json:"id"`
	Name        string     `gorm:"uniqueIndex;not null" json:"name"`
	Description string     `json:"description"`
	Parameters  JSONConfig `gorm:"type:jsonb" json:"parameters"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
