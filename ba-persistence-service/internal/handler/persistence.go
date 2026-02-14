package handler

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/blcvn/backend/services/ba-persistence-service/internal/models"
	"gorm.io/gorm"
)

type PersistenceService struct {
	db *gorm.DB
}

func NewPersistenceService(db *gorm.DB) *PersistenceService {
	return &PersistenceService{db: db}
}

// CaptureInteraction captures an agent interaction
func (s *PersistenceService) CaptureInteraction(interaction *models.AgentInteraction) error {
	// Calculate sizes
	inputBytes, _ := json.Marshal(interaction.Input)
	outputBytes, _ := json.Marshal(interaction.Output)

	interaction.InputSize = len(inputBytes)
	interaction.OutputSize = len(outputBytes)
	interaction.CreatedAt = time.Now()

	// If payload is large (>1MB), store in S3/MinIO
	const maxInlineSize = 1024 * 1024 // 1MB
	if interaction.InputSize+interaction.OutputSize > maxInlineSize {
		// In production, upload to S3/MinIO and store path
		interaction.StoragePath = fmt.Sprintf("s3://ba-agent-data/%s/%d.json",
			interaction.SessionID, time.Now().Unix())

		// Clear inline data
		interaction.Input = models.JSONPayload{"_ref": interaction.StoragePath}
		interaction.Output = models.JSONPayload{"_ref": interaction.StoragePath}
	}

	return s.db.Create(interaction).Error
}

// RecordMetrics records performance metrics
func (s *PersistenceService) RecordMetrics(metrics *models.PerformanceMetrics) error {
	metrics.CreatedAt = time.Now()
	return s.db.Create(metrics).Error
}

// LogError logs an error
func (s *PersistenceService) LogError(errorLog *models.ErrorLog) error {
	errorLog.CreatedAt = time.Now()
	return s.db.Create(errorLog).Error
}

// CreateAuditLog creates an audit log entry
func (s *PersistenceService) CreateAuditLog(log *models.AuditLog) error {
	log.CreatedAt = time.Now()
	return s.db.Create(log).Error
}

// QueryInteractions retrieves interactions with filters
func (s *PersistenceService) QueryInteractions(filters map[string]interface{}, limit int) ([]*models.AgentInteraction, error) {
	var interactions []*models.AgentInteraction

	query := s.db.Model(&models.AgentInteraction{})

	if userID, ok := filters["user_id"].(string); ok {
		query = query.Where("user_id = ?", userID)
	}
	if agentID, ok := filters["agent_id"].(string); ok {
		query = query.Where("agent_id = ?", agentID)
	}
	if sessionID, ok := filters["session_id"].(string); ok {
		query = query.Where("session_id = ?", sessionID)
	}

	err := query.Order("created_at DESC").Limit(limit).Find(&interactions).Error
	return interactions, err
}

// GetInteractionByID retrieves a specific interaction
func (s *PersistenceService) GetInteractionByID(id uint) (*models.AgentInteraction, error) {
	var interaction models.AgentInteraction
	err := s.db.Where("id = ?", id).First(&interaction).Error
	return &interaction, err
}

// GetMetricsForInteraction retrieves metrics for an interaction
func (s *PersistenceService) GetMetricsForInteraction(interactionID uint) (*models.PerformanceMetrics, error) {
	var metrics models.PerformanceMetrics
	err := s.db.Where("interaction_id = ?", interactionID).First(&metrics).Error
	return &metrics, err
}

// GetAuditTrail retrieves audit trail for a user
func (s *PersistenceService) GetAuditTrail(userID string, limit int) ([]*models.AuditLog, error) {
	var logs []*models.AuditLog
	err := s.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

// GetErrorLogs retrieves error logs with filters
func (s *PersistenceService) GetErrorLogs(filters map[string]interface{}, limit int) ([]*models.ErrorLog, error) {
	var errors []*models.ErrorLog

	query := s.db.Model(&models.ErrorLog{})

	if errorType, ok := filters["error_type"].(string); ok {
		query = query.Where("error_type = ?", errorType)
	}

	err := query.Order("created_at DESC").Limit(limit).Find(&errors).Error
	return errors, err
}
