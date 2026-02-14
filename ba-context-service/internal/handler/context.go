package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/blcvn/backend/services/ba-context-service/internal/models"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type ContextService struct {
	db    *gorm.DB
	cache *redis.Client
}

func NewContextService(db *gorm.DB, cache *redis.Client) *ContextService {
	return &ContextService{db: db, cache: cache}
}

// GetUserContext retrieves user context
func (s *ContextService) GetUserContext(ctx context.Context, userID string) (*models.UserContext, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("user_context:%s", userID)
	cached, err := s.cache.Get(ctx, cacheKey).Result()
	if err == nil {
		var uc models.UserContext
		if json.Unmarshal([]byte(cached), &uc) == nil {
			return &uc, nil
		}
	}

	// Fetch from DB
	var uc models.UserContext
	err = s.db.Where("user_id = ?", userID).First(&uc).Error
	if err != nil {
		return nil, err
	}

	// Cache for 5 minutes
	data, _ := json.Marshal(uc)
	s.cache.Set(ctx, cacheKey, data, 5*time.Minute)

	return &uc, nil
}

// GetProjectContext retrieves project context
func (s *ContextService) GetProjectContext(ctx context.Context, projectID string) (*models.ProjectContext, error) {
	var pc models.ProjectContext
	err := s.db.Where("project_id = ?", projectID).First(&pc).Error
	return &pc, err
}

// GetAggregatedContext combines user and project context
func (s *ContextService) GetAggregatedContext(ctx context.Context, userID, projectID string) (*models.AggregatedContext, error) {
	uc, err := s.GetUserContext(ctx, userID)
	if err != nil {
		return nil, err
	}

	pc, err := s.GetProjectContext(ctx, projectID)
	if err != nil {
		return nil, err
	}

	return &models.AggregatedContext{
		UserContext:    uc,
		ProjectContext: pc,
		Timestamp:      time.Now(),
	}, nil
}

// UpdateUserContext updates user context
func (s *ContextService) UpdateUserContext(ctx context.Context, uc *models.UserContext) error {
	err := s.db.Save(uc).Error
	if err != nil {
		return err
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("user_context:%s", uc.UserID)
	s.cache.Del(ctx, cacheKey)

	return nil
}

// UpdateProjectContext updates project context
func (s *ContextService) UpdateProjectContext(ctx context.Context, pc *models.ProjectContext) error {
	pc.Version++
	return s.db.Save(pc).Error
}
