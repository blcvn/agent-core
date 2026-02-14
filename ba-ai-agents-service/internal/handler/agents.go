package handler

import (
	"time"

	"github.com/blcvn/backend/services/ba-ai-agents-service/internal/models"
	"gorm.io/gorm"
)

type AIAgentsService struct {
	db *gorm.DB
}

func NewAIAgentsService(db *gorm.DB) *AIAgentsService {
	return &AIAgentsService{db: db}
}

// CreatePersona creates a new agent persona
func (s *AIAgentsService) CreatePersona(persona *models.AgentPersona) error {
	return s.db.Create(persona).Error
}

// GetPersona retrieves an agent persona by ID
func (s *AIAgentsService) GetPersona(id string) (*models.AgentPersona, error) {
	var persona models.AgentPersona
	err := s.db.Where("id = ? AND is_active = ?", id, true).First(&persona).Error
	return &persona, err
}

// GetPersonaByName retrieves an agent persona by name
func (s *AIAgentsService) GetPersonaByName(name string) (*models.AgentPersona, error) {
	var persona models.AgentPersona
	err := s.db.Where("name = ? AND is_active = ?", name, true).First(&persona).Error
	return &persona, err
}

// ListPersonas returns all active personas
func (s *AIAgentsService) ListPersonas() ([]*models.AgentPersona, error) {
	var personas []*models.AgentPersona
	err := s.db.Where("is_active = ?", true).Find(&personas).Error
	return personas, err
}

// UpdatePersona updates an agent persona
func (s *AIAgentsService) UpdatePersona(persona *models.AgentPersona) error {
	persona.UpdatedAt = time.Now()
	return s.db.Save(persona).Error
}

// DeletePersona soft deletes a persona
func (s *AIAgentsService) DeletePersona(id string) error {
	return s.db.Model(&models.AgentPersona{}).
		Where("id = ?", id).
		Update("is_active", false).Error
}

// RecordTaskCompletion records task completion metrics
func (s *AIAgentsService) RecordTaskCompletion(agentID string, success bool, latencyMs int, tokenUsage int64, costUSD float64, qualityScore float64) error {
	var perf models.AgentPerformance
	err := s.db.Where("agent_id = ?", agentID).First(&perf).Error

	if err == gorm.ErrRecordNotFound {
		// Create new performance record
		perf = models.AgentPerformance{
			AgentID:         agentID,
			TasksCompleted:  1,
			TasksSucceeded:  0,
			TasksFailed:     0,
			AvgQualityScore: qualityScore,
			AvgLatencyMs:    latencyMs,
			TotalTokenUsage: tokenUsage,
			TotalCostUSD:    costUSD,
			LastTaskAt:      time.Now(),
		}
		if success {
			perf.TasksSucceeded = 1
		} else {
			perf.TasksFailed = 1
		}
		return s.db.Create(&perf).Error
	}

	// Update existing record
	perf.TasksCompleted++
	if success {
		perf.TasksSucceeded++
	} else {
		perf.TasksFailed++
	}

	// Update averages
	perf.AvgQualityScore = (perf.AvgQualityScore*float64(perf.TasksCompleted-1) + qualityScore) / float64(perf.TasksCompleted)
	perf.AvgLatencyMs = (perf.AvgLatencyMs*(perf.TasksCompleted-1) + latencyMs) / perf.TasksCompleted
	perf.TotalTokenUsage += tokenUsage
	perf.TotalCostUSD += costUSD
	perf.LastTaskAt = time.Now()
	perf.UpdatedAt = time.Now()

	return s.db.Save(&perf).Error
}

// GetPerformance retrieves performance metrics for an agent
func (s *AIAgentsService) GetPerformance(agentID string) (*models.AgentPerformance, error) {
	var perf models.AgentPerformance
	err := s.db.Where("agent_id = ?", agentID).First(&perf).Error
	return &perf, err
}

// GetTopPerformers returns top N agents by success rate
func (s *AIAgentsService) GetTopPerformers(limit int) ([]*models.AgentPerformance, error) {
	var performers []*models.AgentPerformance
	err := s.db.Order("(tasks_succeeded::float / NULLIF(tasks_completed, 0)) DESC").
		Limit(limit).
		Find(&performers).Error
	return performers, err
}

// AddSpecialization adds domain-specific specialization
func (s *AIAgentsService) AddSpecialization(spec *models.AgentSpecialization) error {
	return s.db.Create(spec).Error
}

// GetSpecializations retrieves all specializations for an agent
func (s *AIAgentsService) GetSpecializations(agentID string) ([]*models.AgentSpecialization, error) {
	var specs []*models.AgentSpecialization
	err := s.db.Where("agent_id = ? AND is_active = ?", agentID, true).Find(&specs).Error
	return specs, err
}
