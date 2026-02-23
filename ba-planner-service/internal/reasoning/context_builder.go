package chat_agent

import (
	"context"

	v32 "github.com/blcvn/ba-shared-libs/pkg/domain/v3.2"
)

// ContextBuilder gathers context for modifications
type ContextBuilder struct {
}

// NewContextBuilder creates a new ContextBuilder
func NewContextBuilder() *ContextBuilder {
	return &ContextBuilder{}
}

// BuildModificationContext gathers related context from KG and Document
func (cb *ContextBuilder) BuildModificationContext(ctx context.Context, affectedSections []AffectedSection, doc *v32.Document, kg *v32.RequirementGraph) (*ModificationContext, error) {
	mCtx := &ModificationContext{
		AffectedSections: affectedSections,
	}

	for _, section := range affectedSections {
		// 1. Gather Related User Stories
		userStories := findRelatedUserStories(kg, section.SectionID)
		mCtx.RelatedUserStories = append(mCtx.RelatedUserStories, userStories...)

		// 2. Gather Related Features
		features := findRelatedFeatures(kg, userStories)
		mCtx.RelatedFeatures = append(mCtx.RelatedFeatures, features...)

		// 3. Gather Related Business Rules
		rules := findRelatedBusinessRules(kg, section.SectionID)
		mCtx.RelatedBusinessRules = append(mCtx.RelatedBusinessRules, rules...)

		// 4. Gather Related Integrations
		integrations := findRelatedIntegrations(kg, section.SectionID)
		mCtx.RelatedIntegrations = append(mCtx.RelatedIntegrations, integrations...)

		// 5. Find Dependent Sections
		deps := findDependentSections(kg, section.SectionID)
		mCtx.DependentSections = append(mCtx.DependentSections, deps...)
	}

	return mCtx, nil
}

func findRelatedUserStories(kg *v32.RequirementGraph, sectionID string) []v32.UserStory {
	// TODO: Graph traversal logic
	return []v32.UserStory{}
}

func findRelatedFeatures(kg *v32.RequirementGraph, userStories []v32.UserStory) []v32.Feature {
	// TODO: Graph traversal logic
	return []v32.Feature{}
}

func findRelatedBusinessRules(kg *v32.RequirementGraph, sectionID string) []v32.BusinessRule {
	// TODO: Graph traversal logic
	return []v32.BusinessRule{}
}

func findRelatedIntegrations(kg *v32.RequirementGraph, sectionID string) []v32.Integration {
	// TODO: Graph traversal logic
	return []v32.Integration{}
}

func findDependentSections(kg *v32.RequirementGraph, sectionID string) []SectionRef {
	// TODO: Graph traversal logic
	return []SectionRef{}
}
