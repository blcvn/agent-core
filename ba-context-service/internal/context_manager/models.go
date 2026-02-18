package context_manager

import (
	"context"
)

// ContextType defines the scope of context
type ContextType string

const (
	UserContext    ContextType = "USER"
	ProjectContext ContextType = "PROJECT"
	GlobalContext  ContextType = "GLOBAL"
)

// ContextEntry represents a single piece of context information
type ContextEntry struct {
	Key       string      `json:"key"`
	Value     interface{} `json:"value"`
	Type      ContextType `json:"type"`
	Timestamp int64       `json:"timestamp"`
}

// ContextManager defines the interface for managing agent context
type ContextManager interface {
	// Set stores a context value
	Set(ctx context.Context, userID string, key string, value interface{}, contextType ContextType) error

	// Get retrieves a context value
	Get(ctx context.Context, userID string, key string, contextType ContextType) (interface{}, error)

	// Clear removes a context value
	Clear(ctx context.Context, userID string, key string, contextType ContextType) error

	// List retrieves all context for a user/scope
	List(ctx context.Context, userID string, contextType ContextType) ([]ContextEntry, error)
}
