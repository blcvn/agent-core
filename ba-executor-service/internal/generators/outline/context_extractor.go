package outline

import (
	"context"
	"fmt"

	v32 "github.com/blcvn/backend/services/pkg/domain/v3.2"
)

// ContextExtractor extracts relevant context from Knowledge Graph for Outline generation
type ContextExtractor struct{}

// NewContextExtractor creates a new context extractor
func NewContextExtractor() *ContextExtractor {
	return &ContextExtractor{}
}

// ModuleContext contains all KG data needed for Outline generation
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
	UseCases      []UseCaseContext
}

// UseCaseContext represents use case information from KG (pre-identified in Index phase)
type UseCaseContext struct {
	ID           string
	Name         string
	Description  string
	PrimaryActor string
	Trigger      string
	Priority     string
	RelatedUS    []string // Related User Story IDs
}

// FeatureContext represents feature information from KG
type FeatureContext struct {
	ID          string
	Name        string
	Description string
	Priority    string
}

// UserStoryContext represents user story information from KG
type UserStoryContext struct {
	ID        string
	AsA       string
	IWant     string
	SoThat    string
	FeatureID string
}

// PersonaContext represents persona information from KG
type PersonaContext struct {
	ID          string
	Name        string
	Role        string
	Description string
	Goals       []string
	PainPoints  []string
}

// IntegrationContext represents integration information from KG
type IntegrationContext struct {
	ID         string
	SystemName string
	Type       string
	Purpose    string
	Direction  string
}

// BusinessRuleContext represents business rule information from KG
type BusinessRuleContext struct {
	ID          string
	Name        string
	Description string
}

// EntityContext represents data entity information from KG
type EntityContext struct {
	ID          string
	Name        string
	Description string
	Attributes  []string
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
		UseCases:      []UseCaseContext{},
	}

	// Extract product information from KG metadata
	if name, ok := kg.Metadata["product_name"].(string); ok {
		moduleCtx.ProductName = name
	}
	if vision, ok := kg.Metadata["vision"].(string); ok {
		moduleCtx.ProductVision = vision
	}

	// Extract data from KG nodes
	for _, node := range kg.Nodes {
		switch node.Type {
		case v32.ReqTypeFunctional:
			moduleCtx.Features = append(moduleCtx.Features, FeatureContext{
				ID:          node.ReferenceID,
				Name:        node.Summary,
				Description: node.Description,
				Priority:    e.getMetadataString(node, "priority"),
			})

		case v32.ReqTypeUserStory:
			moduleCtx.UserStories = append(moduleCtx.UserStories, UserStoryContext{
				ID:        node.ReferenceID,
				AsA:       e.getMetadataString(node, "as_a"),
				IWant:     node.Summary,
				SoThat:    node.Description,
				FeatureID: e.getMetadataString(node, "feature_id"),
			})

		case v32.ReqTypeEntity:
			moduleCtx.Entities = append(moduleCtx.Entities, EntityContext{
				ID:          node.ReferenceID,
				Name:        node.Summary,
				Description: node.Description,
				Attributes:  e.getMetadataStringSlice(node, "attributes"),
			})

		case v32.ReqTypePersona:
			moduleCtx.Personas = append(moduleCtx.Personas, PersonaContext{
				ID:          node.ReferenceID,
				Name:        node.Summary,
				Role:        e.getMetadataString(node, "role"),
				Description: node.Description,
				Goals:       e.getMetadataStringSlice(node, "goals"),
				PainPoints:  e.getMetadataStringSlice(node, "pain_points"),
			})

		case v32.ReqTypeAPI:
			moduleCtx.Integrations = append(moduleCtx.Integrations, IntegrationContext{
				ID:         node.ReferenceID,
				SystemName: node.Summary,
				Type:       e.getMetadataString(node, "type"),
				Purpose:    node.Description,
				Direction:  e.getMetadataString(node, "direction"),
			})

		case v32.ReqTypeBusinessRule:
			moduleCtx.BusinessRules = append(moduleCtx.BusinessRules, BusinessRuleContext{
				ID:          node.ReferenceID,
				Name:        node.Summary,
				Description: node.Description,
			})

		case v32.ReqTypeUseCase:
			// Find related user stories via edges
			relatedUS := []string{}
			for _, edge := range kg.Edges {
				if edge.TargetID == node.ID && edge.Type == v32.DepTypeRefines {
					relatedUS = append(relatedUS, edge.SourceID)
				}
			}

			moduleCtx.UseCases = append(moduleCtx.UseCases, UseCaseContext{
				ID:           node.ReferenceID,
				Name:         node.Summary,
				Description:  node.Description,
				PrimaryActor: e.getMetadataString(node, "primary_actor"),
				Trigger:      e.getMetadataString(node, "trigger"),
				Priority:     e.getMetadataString(node, "priority"),
				RelatedUS:    relatedUS,
			})
		}
	}

	return moduleCtx, nil
}

func (e *ContextExtractor) getMetadataString(node v32.RequirementNode, key string) string {
	if val, ok := node.Metadata[key].(string); ok {
		return val
	}
	return ""
}

func (e *ContextExtractor) getMetadataStringSlice(node v32.RequirementNode, key string) []string {
	if val, ok := node.Metadata[key].([]interface{}); ok {
		result := []string{}
		for _, v := range val {
			if s, ok := v.(string); ok {
				result = append(result, s)
			}
		}
		return result
	}
	return []string{}
}
