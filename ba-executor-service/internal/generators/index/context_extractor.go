package index

import (
	"context"
	"fmt"
	"strings"

	v32 "github.com/blcvn/ba-shared-libs/pkg/domain/v3.2"
)

// ContextExtractor extracts relevant context from Knowledge Graph for Index generation
type ContextExtractor struct{}

// NewContextExtractor creates a new context extractor
func NewContextExtractor() *ContextExtractor {
	return &ContextExtractor{}
}

// ModuleContext contains all KG data needed for Index generation
type ModuleContext struct {
	ModuleName    string
	ProductName   string
	ProductVision string
	Features      []FeatureContext
	UserStories   []UserStoryContext
	Personas      []PersonaContext
	Integrations  []IntegrationContext
	BusinessRules []BusinessRuleContext
	Entities      []EntityContext
}

// FeatureContext represents feature information from KG
type FeatureContext struct {
	ID          string
	Name        string
	Description string
	Priority    string
	Status      string
	Category    string
}

// UserStoryContext represents user story information from KG
type UserStoryContext struct {
	ID        string
	AsA       string
	IWant     string
	SoThat    string
	FeatureID string
	Priority  string
	Size      string
	Status    string
}

// PersonaContext represents persona information from KG
type PersonaContext struct {
	ID             string
	Name           string
	Role           string
	Goals          []string
	PainPoints     []string
	TechnicalLevel string
}

// IntegrationContext represents integration information from KG
type IntegrationContext struct {
	ID         string
	SystemName string
	Type       string
	Purpose    string
	Direction  string
	Status     string
}

// BusinessRuleContext represents business rule information from KG
type BusinessRuleContext struct {
	ID          string
	Name        string
	Description string
	AppliesTo   []string
	Severity    string
}

// EntityContext represents data entity information from KG
type EntityContext struct {
	ID          string
	Name        string
	Description string
	Attributes  []string
}

// cleanID removes the UUID prefix from an ID if present
// Example: "27baf50e-7f1c-48ea-92d8-ac6fa8251af4_F001" -> "F001"
func (e *ContextExtractor) cleanID(id string) string {
	if id == "" {
		return ""
	}
	parts := strings.Split(id, "_")
	if len(parts) > 1 {
		// Return the last part as the ID (e.g., F001, US-001)
		return parts[len(parts)-1]
	}
	return id
}

// ExtractModuleContext extracts all relevant context from KG for a specific module
func (e *ContextExtractor) ExtractModuleContext(ctx context.Context, kg *v32.RequirementGraph, moduleName string) (*ModuleContext, error) {
	if kg == nil {
		return nil, fmt.Errorf("knowledge graph is nil")
	}

	moduleCtx := &ModuleContext{
		ModuleName:    moduleName,
		Features:      []FeatureContext{},
		UserStories:   []UserStoryContext{},
		Personas:      []PersonaContext{},
		Integrations:  []IntegrationContext{},
		BusinessRules: []BusinessRuleContext{},
		Entities:      []EntityContext{},
	}

	// Extract product information from KG metadata
	if name, ok := kg.Metadata["product_name"].(string); ok {
		moduleCtx.ProductName = name
	}
	if vision, ok := kg.Metadata["vision"].(string); ok {
		moduleCtx.ProductVision = vision
	}

	// Extract features from KG nodes
	for _, node := range kg.Nodes {
		switch node.Type {
		case v32.ReqTypeFunctional:
			feature := e.extractFeatureFromNode(node)
			moduleCtx.Features = append(moduleCtx.Features, feature)

		case v32.ReqTypeUserStory:
			userStory := e.extractUserStoryFromNode(node)
			moduleCtx.UserStories = append(moduleCtx.UserStories, userStory)

		case v32.ReqTypeEntity:
			entity := e.extractEntityFromNode(node)
			moduleCtx.Entities = append(moduleCtx.Entities, entity)

		case v32.ReqTypePersona:
			persona := e.extractPersonaFromNode(node)
			moduleCtx.Personas = append(moduleCtx.Personas, persona)

		case v32.ReqTypeAPI:
			integration := e.extractIntegrationFromNode(node)
			moduleCtx.Integrations = append(moduleCtx.Integrations, integration)

		case v32.ReqTypeBusinessRule:
			rule := e.extractBusinessRuleFromNode(node)
			moduleCtx.BusinessRules = append(moduleCtx.BusinessRules, rule)
		}
	}

	return moduleCtx, nil
}

// Helper functions to extract specific node types

func (e *ContextExtractor) extractFeatureFromNode(node v32.RequirementNode) FeatureContext {
	feature := FeatureContext{
		ID: node.ReferenceID,
	}

	feature.Name = node.Summary
	feature.Description = node.Description

	if priority, ok := node.Metadata["priority"].(string); ok {
		feature.Priority = priority
	}
	if status, ok := node.Metadata["status"].(string); ok {
		feature.Status = status
	}
	if category, ok := node.Metadata["category"].(string); ok {
		feature.Category = category
	}

	return feature
}

func (e *ContextExtractor) extractUserStoryFromNode(node v32.RequirementNode) UserStoryContext {
	story := UserStoryContext{
		ID: node.ReferenceID,
	}

	story.IWant = node.Summary
	story.SoThat = node.Description // Simplified mapping

	if asA, ok := node.Metadata["as_a"].(string); ok {
		story.AsA = asA
	}
	if featureID, ok := node.Metadata["feature_id"].(string); ok {
		story.FeatureID = e.cleanID(featureID)
	}
	if priority, ok := node.Metadata["priority"].(string); ok {
		story.Priority = priority
	}
	if size, ok := node.Metadata["size"].(string); ok {
		story.Size = size
	}
	if status, ok := node.Metadata["status"].(string); ok {
		story.Status = status
	}

	return story
}

func (e *ContextExtractor) extractPersonaFromNode(node v32.RequirementNode) PersonaContext {
	persona := PersonaContext{
		ID:         node.ReferenceID,
		Goals:      []string{},
		PainPoints: []string{},
	}

	persona.Name = node.Summary

	if role, ok := node.Metadata["role"].(string); ok {
		persona.Role = role
	}
	if techLevel, ok := node.Metadata["technical_level"].(string); ok {
		persona.TechnicalLevel = techLevel
	}

	// Extract arrays from Metadata
	if goals, ok := node.Metadata["goals"].([]interface{}); ok {
		for _, g := range goals {
			if goal, ok := g.(string); ok {
				persona.Goals = append(persona.Goals, goal)
			}
		}
	}
	if painPoints, ok := node.Metadata["pain_points"].([]interface{}); ok {
		for _, p := range painPoints {
			if point, ok := p.(string); ok {
				persona.PainPoints = append(persona.PainPoints, point)
			}
		}
	}

	return persona
}

func (e *ContextExtractor) extractIntegrationFromNode(node v32.RequirementNode) IntegrationContext {
	integration := IntegrationContext{
		ID: node.ReferenceID,
	}

	integration.SystemName = node.Summary

	if intType, ok := node.Metadata["type"].(string); ok {
		integration.Type = intType
	}
	if purpose, ok := node.Metadata["purpose"].(string); ok {
		integration.Purpose = purpose
	}
	if direction, ok := node.Metadata["direction"].(string); ok {
		integration.Direction = direction
	}
	if status, ok := node.Metadata["status"].(string); ok {
		integration.Status = status
	}

	return integration
}

func (e *ContextExtractor) extractBusinessRuleFromNode(node v32.RequirementNode) BusinessRuleContext {
	rule := BusinessRuleContext{
		ID:        node.ReferenceID,
		AppliesTo: []string{},
	}

	rule.Name = node.Summary
	rule.Description = node.Description

	if severity, ok := node.Metadata["severity"].(string); ok {
		rule.Severity = severity
	}

	// Extract applies_to array
	if appliesTo, ok := node.Metadata["applies_to"].([]interface{}); ok {
		for _, a := range appliesTo {
			if target, ok := a.(string); ok {
				rule.AppliesTo = append(rule.AppliesTo, target)
			}
		}
	}

	return rule
}

func (e *ContextExtractor) extractEntityFromNode(node v32.RequirementNode) EntityContext {
	entity := EntityContext{
		ID:         node.ReferenceID,
		Attributes: []string{},
	}

	entity.Name = node.Summary
	entity.Description = node.Description

	// Extract attributes
	if attrs, ok := node.Metadata["attributes"].([]interface{}); ok {
		for _, a := range attrs {
			if attr, ok := a.(string); ok {
				entity.Attributes = append(entity.Attributes, attr)
			}
		}
	}

	return entity
}
