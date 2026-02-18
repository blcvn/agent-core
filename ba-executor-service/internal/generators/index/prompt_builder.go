package index

import (
	"fmt"
	"strings"
)

// PromptBuilder builds prompts for URD Index generation
type PromptBuilder struct{}

// NewPromptBuilder creates a new prompt builder
func NewPromptBuilder() *PromptBuilder {
	return &PromptBuilder{}
}

// BuildUserPrompt creates the user prompt with context and dynamic instruction
func (pb *PromptBuilder) BuildUserPrompt(ctx *ModuleContext, instructionTemplate string) string {
	var sb strings.Builder

	// Header
	sb.WriteString("Generate a URD Index document for the following module.\n")
	sb.WriteString("This is the 'Skeleton' phase. You must generate the structure and the list of Use Case headers (Section 5) based on the User Stories.\n")
	sb.WriteString("For other sections (1, 2, 3, 4, 6, 7, 8), generate ONLY the static titles/headers as defined in the template.\n\n")

	sb.WriteString(fmt.Sprintf("**Module Name:** %s\n", ctx.ModuleName))
	sb.WriteString(fmt.Sprintf("**Product:** %s\n", ctx.ProductName))
	if ctx.ProductVision != "" {
		sb.WriteString(fmt.Sprintf("**Vision:** %s\n", ctx.ProductVision))
	}
	sb.WriteString("\n---\n\n")

	// Context sections
	pb.writePersonasContext(&sb, ctx)
	pb.writeFeaturesContext(&sb, ctx)
	pb.writeUserStoriesContext(&sb, ctx)
	pb.writeIntegrationsContext(&sb, ctx)
	pb.writeBusinessRulesContext(&sb, ctx)
	pb.writeEntitiesContext(&sb, ctx)

	// Output format specification
	sb.WriteString("\n---\n\n")
	sb.WriteString("# OUTPUT FORMAT\n\n")
	sb.WriteString("Generate a URD Index document following this EXACT structure.\n")
	sb.WriteString("IMPORTANT: Do NOT generate data tables. Do NOT generate detailed flows. ONLY generate the structure and Section 5 headers.\n\n")
	sb.WriteString(instructionTemplate)

	return sb.String()
}

func (pb *PromptBuilder) writePersonasContext(sb *strings.Builder, ctx *ModuleContext) {
	sb.WriteString("# CONTEXT: Personas\n\n")
	if len(ctx.Personas) == 0 {
		sb.WriteString("No personas defined.\n\n")
		return
	}

	for _, persona := range ctx.Personas {
		sb.WriteString(fmt.Sprintf("## %s - %s\n", persona.ID, persona.Name))
		sb.WriteString(fmt.Sprintf("- **Role:** %s\n", persona.Role))
		sb.WriteString(fmt.Sprintf("- **Technical Level:** %s\n", persona.TechnicalLevel))
		if len(persona.Goals) > 0 {
			sb.WriteString("- **Goals:**\n")
			for _, goal := range persona.Goals {
				sb.WriteString(fmt.Sprintf("  - %s\n", goal))
			}
		}
		if len(persona.PainPoints) > 0 {
			sb.WriteString("- **Pain Points:**\n")
			for _, pain := range persona.PainPoints {
				sb.WriteString(fmt.Sprintf("  - %s\n", pain))
			}
		}
		sb.WriteString("\n")
	}
}

func (pb *PromptBuilder) writeFeaturesContext(sb *strings.Builder, ctx *ModuleContext) {
	sb.WriteString("# CONTEXT: Features\n\n")
	if len(ctx.Features) == 0 {
		sb.WriteString("No features defined.\n\n")
		return
	}

	for _, feature := range ctx.Features {
		sb.WriteString(fmt.Sprintf("## %s - %s\n", feature.ID, feature.Name))
		sb.WriteString(fmt.Sprintf("- **Description:** %s\n", feature.Description))
		sb.WriteString(fmt.Sprintf("- **Priority:** %s\n", feature.Priority))
		if feature.Status != "" {
			sb.WriteString(fmt.Sprintf("- **Status:** %s\n", feature.Status))
		}
		if feature.Category != "" {
			sb.WriteString(fmt.Sprintf("- **Category:** %s\n", feature.Category))
		}
		sb.WriteString("\n")
	}
}

func (pb *PromptBuilder) writeUserStoriesContext(sb *strings.Builder, ctx *ModuleContext) {
	sb.WriteString("# CONTEXT: User Stories\n\n")
	if len(ctx.UserStories) == 0 {
		sb.WriteString("No user stories defined.\n\n")
		return
	}

	for _, story := range ctx.UserStories {
		sb.WriteString(fmt.Sprintf("## %s\n", story.ID))
		sb.WriteString(fmt.Sprintf("- **As a:** %s\n", story.AsA))
		sb.WriteString(fmt.Sprintf("- **I want:** %s\n", story.IWant))
		sb.WriteString(fmt.Sprintf("- **So that:** %s\n", story.SoThat))
		sb.WriteString(fmt.Sprintf("- **Feature:** %s\n", story.FeatureID))
		sb.WriteString(fmt.Sprintf("- **Priority:** %s\n", story.Priority))
		if story.Size != "" {
			sb.WriteString(fmt.Sprintf("- **Size:** %s\n", story.Size))
		}
		if story.Status != "" {
			sb.WriteString(fmt.Sprintf("- **Status:** %s\n", story.Status))
		}
		sb.WriteString("\n")
	}
}

func (pb *PromptBuilder) writeIntegrationsContext(sb *strings.Builder, ctx *ModuleContext) {
	sb.WriteString("# CONTEXT: Integrations\n\n")
	if len(ctx.Integrations) == 0 {
		sb.WriteString("No integrations defined.\n\n")
		return
	}

	for _, integration := range ctx.Integrations {
		sb.WriteString(fmt.Sprintf("## %s - %s\n", integration.ID, integration.SystemName))
		sb.WriteString(fmt.Sprintf("- **Type:** %s\n", integration.Type))
		sb.WriteString(fmt.Sprintf("- **Purpose:** %s\n", integration.Purpose))
		sb.WriteString(fmt.Sprintf("- **Direction:** %s\n", integration.Direction))
		if integration.Status != "" {
			sb.WriteString(fmt.Sprintf("- **Status:** %s\n", integration.Status))
		}
		sb.WriteString("\n")
	}
}

func (pb *PromptBuilder) writeBusinessRulesContext(sb *strings.Builder, ctx *ModuleContext) {
	sb.WriteString("# CONTEXT: Business Rules\n\n")
	if len(ctx.BusinessRules) == 0 {
		sb.WriteString("No business rules defined.\n\n")
		return
	}

	for _, rule := range ctx.BusinessRules {
		sb.WriteString(fmt.Sprintf("## %s - %s\n", rule.ID, rule.Name))
		sb.WriteString(fmt.Sprintf("- **Description:** %s\n", rule.Description))
		if rule.Severity != "" {
			sb.WriteString(fmt.Sprintf("- **Severity:** %s\n", rule.Severity))
		}
		if len(rule.AppliesTo) > 0 {
			sb.WriteString("- **Applies To:** ")
			sb.WriteString(strings.Join(rule.AppliesTo, ", "))
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}
}

func (pb *PromptBuilder) writeEntitiesContext(sb *strings.Builder, ctx *ModuleContext) {
	sb.WriteString("# CONTEXT: Data Entities\n\n")
	if len(ctx.Entities) == 0 {
		sb.WriteString("No data entities defined.\n\n")
		return
	}

	for _, entity := range ctx.Entities {
		sb.WriteString(fmt.Sprintf("## %s - %s\n", entity.ID, entity.Name))
		if entity.Description != "" {
			sb.WriteString(fmt.Sprintf("- **Description:** %s\n", entity.Description))
		}
		if len(entity.Attributes) > 0 {
			sb.WriteString("- **Key Attributes:**\n")
			for _, attr := range entity.Attributes {
				sb.WriteString(fmt.Sprintf("  - %s\n", attr))
			}
		}
		sb.WriteString("\n")
	}
}
