package outline

import (
	"fmt"
	"strings"
)

// PromptBuilder builds prompts for URD Outline generation
type PromptBuilder struct{}

// NewPromptBuilder creates a new prompt builder
func NewPromptBuilder() *PromptBuilder {
	return &PromptBuilder{}
}

// BuildUserPrompt creates the user prompt with context and dynamic instruction
func (pb *PromptBuilder) BuildUserPrompt(ctx *ModuleContext, instructionTemplate string, previousContent string) string {
	var sb strings.Builder

	// Header
	sb.WriteString("Generate a detailed URD Outline document for the following module.\n")
	sb.WriteString("This is the 'Outline' phase. You must FILL IN all the tables defined in the structure below with specific data from the Context.\n")

	if previousContent != "" {
		sb.WriteString("You are provided with the 'URD Index' (Skeleton) below. You must PRESERVE the existing structure and ONLY fill in the missing details/tables.\n")
		sb.WriteString("--- URD INDEX (PREVIOUS STEP) ---\n")
		sb.WriteString(previousContent)
		sb.WriteString("\n--- END URD INDEX ---\n\n")
	} else {
		sb.WriteString("If Use Cases are not pre-identified in the context, you must identify them from the User Stories (1:1 or 1:n mapping).\n\n")
	}

	sb.WriteString(fmt.Sprintf("**Module Name:** %s\n", ctx.ModuleName))
	sb.WriteString(fmt.Sprintf("**Product:** %s\n", ctx.ProductName))
	if ctx.ProductVision != "" {
		sb.WriteString(fmt.Sprintf("**Vision:** %s\n", ctx.ProductVision))
	}
	sb.WriteString("\n---\n\n")

	// Context sections
	pb.writeUseCasesContext(&sb, ctx)
	pb.writePersonasContext(&sb, ctx)
	pb.writeFeaturesContext(&sb, ctx)
	pb.writeUserStoriesContext(&sb, ctx)
	pb.writeIntegrationsContext(&sb, ctx)
	pb.writeBusinessRulesContext(&sb, ctx)
	pb.writeEntitiesContext(&sb, ctx)

	// Output format specification
	sb.WriteString("\n---\n\n")
	sb.WriteString("# OUTPUT FORMAT\n\n")
	sb.WriteString("Generate a complete URD Outline document following this EXACT structure.\n")
	sb.WriteString("IMPORTANT: Ensure all tables are filled with logical data derived from the User Stories and Context.\n\n")
	sb.WriteString(instructionTemplate)

	return sb.String()
}

func (pb *PromptBuilder) writeUseCasesContext(sb *strings.Builder, ctx *ModuleContext) {
	sb.WriteString("# CONTEXT: Pre-Identified Use Cases\n\n")
	if len(ctx.UseCases) == 0 {
		sb.WriteString("No use cases identified in the previous step.\n\n")
		return
	}

	for _, uc := range ctx.UseCases {
		sb.WriteString(fmt.Sprintf("## %s - %s\n", uc.ID, uc.Name))
		sb.WriteString(fmt.Sprintf("- **Summary:** %s\n", uc.Description))
		sb.WriteString(fmt.Sprintf("- **Primary Actor:** %s\n", uc.PrimaryActor))
		sb.WriteString(fmt.Sprintf("- **Trigger:** %s\n", uc.Trigger))
		sb.WriteString(fmt.Sprintf("- **Priority:** %s\n", uc.Priority))
		if len(uc.RelatedUS) > 0 {
			sb.WriteString(fmt.Sprintf("- **Based on User Stories:** %s\n", strings.Join(uc.RelatedUS, ", ")))
		}
		sb.WriteString("\n")
	}
}

func (pb *PromptBuilder) writePersonasContext(sb *strings.Builder, ctx *ModuleContext) {
	sb.WriteString("# CONTEXT: Personas/Actors\n\n")
	for _, p := range ctx.Personas {
		sb.WriteString(fmt.Sprintf("## %s - %s\n", p.ID, p.Name))
		sb.WriteString(fmt.Sprintf("- **Role:** %s\n", p.Role))
		sb.WriteString(fmt.Sprintf("- **Description:** %s\n", p.Description))
		sb.WriteString("\n")
	}
}

func (pb *PromptBuilder) writeFeaturesContext(sb *strings.Builder, ctx *ModuleContext) {
	sb.WriteString("# CONTEXT: Features\n\n")
	for _, f := range ctx.Features {
		sb.WriteString(fmt.Sprintf("- %s: %s (%s)\n", f.ID, f.Name, f.Description))
	}
	sb.WriteString("\n")
}

func (pb *PromptBuilder) writeUserStoriesContext(sb *strings.Builder, ctx *ModuleContext) {
	sb.WriteString("# CONTEXT: User Stories\n\n")
	for _, s := range ctx.UserStories {
		sb.WriteString(fmt.Sprintf("- %s: As a %s, I want %s, so that %s\n", s.ID, s.AsA, s.IWant, s.SoThat))
	}
	sb.WriteString("\n")
}

func (pb *PromptBuilder) writeIntegrationsContext(sb *strings.Builder, ctx *ModuleContext) {
	sb.WriteString("# CONTEXT: Integrations\n\n")
	for _, i := range ctx.Integrations {
		sb.WriteString(fmt.Sprintf("- %s (%s): %s (%s)\n", i.ID, i.SystemName, i.Purpose, i.Direction))
	}
	sb.WriteString("\n")
}

func (pb *PromptBuilder) writeBusinessRulesContext(sb *strings.Builder, ctx *ModuleContext) {
	sb.WriteString("# CONTEXT: Business Rules\n\n")
	for _, r := range ctx.BusinessRules {
		sb.WriteString(fmt.Sprintf("- %s: %s - %s\n", r.ID, r.Name, r.Description))
	}
	sb.WriteString("\n")
}

func (pb *PromptBuilder) writeEntitiesContext(sb *strings.Builder, ctx *ModuleContext) {
	sb.WriteString("# CONTEXT: Data Entities\n\n")
	for _, e := range ctx.Entities {
		sb.WriteString(fmt.Sprintf("- %s: %s - %s\n", e.ID, e.Name, e.Description))
	}
	sb.WriteString("\n")
}
