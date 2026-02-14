package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/blcvn/backend/services/ba-state-service/internal/models"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type StateManager struct {
	db    *gorm.DB
	cache *redis.Client
}

func NewStateManager(db *gorm.DB, cache *redis.Client) *StateManager {
	return &StateManager{db: db, cache: cache}
}

// SaveState saves session state (Redis for fast access, Postgres for persistence)
func (m *StateManager) SaveState(ctx context.Context, state *models.SessionState) error {
	// Save to Postgres
	if err := m.db.Save(state).Error; err != nil {
		return err
	}

	// Cache in Redis
	cacheKey := fmt.Sprintf("session_state:%s", state.ID)
	data, _ := json.Marshal(state)
	return m.cache.Set(ctx, cacheKey, data, 30*time.Minute).Err()
}

// GetState retrieves session state
func (m *StateManager) GetState(ctx context.Context, sessionID string) (*models.SessionState, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("session_state:%s", sessionID)
	cached, err := m.cache.Get(ctx, cacheKey).Result()
	if err == nil {
		var state models.SessionState
		if json.Unmarshal([]byte(cached), &state) == nil {
			return &state, nil
		}
	}

	// Fetch from DB
	var state models.SessionState
	err = m.db.Where("id = ?", sessionID).First(&state).Error
	if err != nil {
		return nil, err
	}

	// Re-cache
	data, _ := json.Marshal(state)
	m.cache.Set(ctx, cacheKey, data, 30*time.Minute)

	return &state, nil
}

// TransitionState transitions state based on event
func (m *StateManager) TransitionState(ctx context.Context, sessionID string, event string) (*models.SessionState, error) {
	state, err := m.GetState(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// Simple state machine logic
	newState := m.computeNextState(state.State, event)
	state.State = newState
	state.UpdatedAt = time.Now()

	if err := m.SaveState(ctx, state); err != nil {
		return nil, err
	}

	return state, nil
}

func (m *StateManager) computeNextState(current models.WorkflowState, event string) models.WorkflowState {
	transitions := map[models.WorkflowState]map[string]models.WorkflowState{
		models.StatePending: {
			"start": models.StateRunning,
		},
		models.StateRunning: {
			"complete":   models.StateCompleted,
			"wait_input": models.StateWaitingForInput,
			"fail":       models.StateFailed,
		},
		models.StateWaitingForInput: {
			"resume": models.StateResumed,
		},
		models.StateFailed: {
			"retry": models.StateRetrying,
		},
		models.StateRetrying: {
			"start": models.StateRunning,
		},
	}

	if nextStates, ok := transitions[current]; ok {
		if next, ok := nextStates[event]; ok {
			return next
		}
	}

	return current
}

// CreateSnapshot creates a checkpoint
func (m *StateManager) CreateSnapshot(ctx context.Context, sessionID string, data models.StateData) error {
	// Get current version
	var count int64
	m.db.Model(&models.StateSnapshot{}).Where("session_id = ?", sessionID).Count(&count)

	snapshot := &models.StateSnapshot{
		SessionID: sessionID,
		Snapshot:  data,
		Version:   int(count) + 1,
		CreatedAt: time.Now(),
	}

	return m.db.Create(snapshot).Error
}

// RestoreSnapshot restores from a checkpoint
func (m *StateManager) RestoreSnapshot(ctx context.Context, sessionID string, version int) (*models.StateSnapshot, error) {
	var snapshot models.StateSnapshot
	err := m.db.Where("session_id = ? AND version = ?", sessionID, version).First(&snapshot).Error
	return &snapshot, err
}
