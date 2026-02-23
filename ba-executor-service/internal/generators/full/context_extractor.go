package full

import (
	"context"
	"fmt"
	"strings"

	v32 "github.com/blcvn/ba-shared-libs/pkg/domain/v3.2"
)

// ContextExtractor extracts comprehensive context from Knowledge Graph for Full URD generation
type ContextExtractor struct{}

// NewContextExtractor creates a new context extractor
func NewContextExtractor() *ContextExtractor {
	return &ContextExtractor{}
}

// ModuleContext contains the complete set of KG data for final URD synthesis
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

// UseCaseContext represents use case information from KG (fully detailed)
type UseCaseContext struct {
	ID             string
	Name           string
	Description    string
	PrimaryActor   string
	Trigger        string
	Preconditions  []string
	Postconditions []string
	MainFlow       []string
	Priority       string
}

// FeatureContext represents feature information from KG
type FeatureContext struct {
	ID          string
	Name        string
	Description string
}

// UserStoryContext represents user story information from KG
type UserStoryContext struct {
	ID     string
	AsA    string
	IWant  string
	SoThat string
}

// PersonaContext represents persona information from KG
type PersonaContext struct {
	ID          string
	Name        string
	Role        string
	Description string
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
			})

		case v32.ReqTypeUserStory:
			moduleCtx.UserStories = append(moduleCtx.UserStories, UserStoryContext{
				ID:     node.ReferenceID,
				AsA:    e.getMetadataString(node, "as_a"),
				IWant:  node.Summary,
				SoThat: node.Description,
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
			moduleCtx.UseCases = append(moduleCtx.UseCases, UseCaseContext{
				ID:             node.ReferenceID,
				Name:           node.Summary,
				Description:    node.Description,
				PrimaryActor:   e.getMetadataString(node, "primary_actor"),
				Trigger:        e.getMetadataString(node, "trigger"),
				Preconditions:  e.getMetadataStringSlice(node, "preconditions"),
				Postconditions: e.getMetadataStringSlice(node, "postconditions"),
				MainFlow:       e.getMetadataStringSlice(node, "main_flow"),
				Priority:       e.getMetadataString(node, "priority"),
			})
		}
	}

	return moduleCtx, nil
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
	} else if val, ok := node.Metadata[key].([]string); ok {
		return val
	}
	return []string{}
}
