package repository

import (
	"context"

	"github.com/blcvn/backend/services/ba-ai-agents-service/internal/models"
)

type AgentRepository interface {
	GetAgent(ctx context.Context, id string) (*models.Agent, error)
	ListAgents(ctx context.Context) ([]*models.Agent, error)
	SaveAgent(ctx context.Context, agent *models.Agent) error
}

type SkillRepository interface {
	GetSkill(ctx context.Context, id string) (*models.Skill, error)
	ListSkills(ctx context.Context) ([]*models.Skill, error)
	SaveSkill(ctx context.Context, skill *models.Skill) error
}
