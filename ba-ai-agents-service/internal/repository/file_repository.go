package repository

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/blcvn/backend/services/ba-ai-agents-service/internal/models"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

// FileAgentRepository implements AgentRepository using YAML files
type FileAgentRepository struct {
	mu     sync.RWMutex
	agents map[string]*models.Agent
}

func NewFileAgentRepository() *FileAgentRepository {
	return &FileAgentRepository{
		agents: make(map[string]*models.Agent),
	}
}

func (r *FileAgentRepository) GetAgent(ctx context.Context, id string) (*models.Agent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agent, ok := r.agents[id]
	if !ok {
		return nil, fmt.Errorf("agent not found: %s", id)
	}
	return agent, nil
}

func (r *FileAgentRepository) ListAgents(ctx context.Context) ([]*models.Agent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var agents []*models.Agent
	for _, a := range r.agents {
		agents = append(agents, a)
	}
	return agents, nil
}

func (r *FileAgentRepository) SaveAgent(ctx context.Context, agent *models.Agent) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if agent.ID == "" {
		agent.ID = uuid.New().String()
	}
	agent.UpdatedAt = time.Now()
	if agent.CreatedAt.IsZero() {
		agent.CreatedAt = time.Now()
	}
	r.agents[agent.ID] = agent
	return nil
}

// LoadFromDirectory reads YAML agent definitions from a directory
func (r *FileAgentRepository) LoadFromDirectory(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || filepath.Ext(path) != ".yaml" {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", path, err)
		}

		var agent models.Agent
		if err := yaml.Unmarshal(content, &agent); err != nil {
			return fmt.Errorf("failed to parse %s: %w", path, err)
		}

		if agent.ID == "" {
			agent.ID = uuid.New().String()
		}
		if agent.CreatedAt.IsZero() {
			agent.CreatedAt = time.Now()
		}
		agent.UpdatedAt = time.Now()
		agent.IsActive = true

		r.mu.Lock()
		r.agents[agent.ID] = &agent
		r.mu.Unlock()

		return nil
	})
}

// FileSkillRepository implements SkillRepository using YAML files
type FileSkillRepository struct {
	mu     sync.RWMutex
	skills map[string]*models.Skill
}

func NewFileSkillRepository() *FileSkillRepository {
	return &FileSkillRepository{
		skills: make(map[string]*models.Skill),
	}
}

func (r *FileSkillRepository) GetSkill(ctx context.Context, id string) (*models.Skill, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	skill, ok := r.skills[id]
	if !ok {
		return nil, fmt.Errorf("skill not found: %s", id)
	}
	return skill, nil
}

func (r *FileSkillRepository) ListSkills(ctx context.Context) ([]*models.Skill, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var skills []*models.Skill
	for _, s := range r.skills {
		skills = append(skills, s)
	}
	return skills, nil
}

func (r *FileSkillRepository) SaveSkill(ctx context.Context, skill *models.Skill) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if skill.ID == "" {
		skill.ID = uuid.New().String()
	}
	skill.UpdatedAt = time.Now()
	if skill.CreatedAt.IsZero() {
		skill.CreatedAt = time.Now()
	}
	r.skills[skill.ID] = skill
	return nil
}

// LoadFromDirectory reads YAML skill definitions from a directory
func (r *FileSkillRepository) LoadFromDirectory(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || filepath.Ext(path) != ".yaml" {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", path, err)
		}

		var skill models.Skill
		if err := yaml.Unmarshal(content, &skill); err != nil {
			return fmt.Errorf("failed to parse %s: %w", path, err)
		}

		if skill.ID == "" {
			skill.ID = uuid.New().String()
		}
		if skill.CreatedAt.IsZero() {
			skill.CreatedAt = time.Now()
		}
		skill.UpdatedAt = time.Now()

		r.mu.Lock()
		r.skills[skill.ID] = &skill
		r.mu.Unlock()

		return nil
	})
}
