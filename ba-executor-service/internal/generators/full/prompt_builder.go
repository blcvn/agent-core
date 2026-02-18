package full

import (
	"fmt"
	"strings"
)

// PromptBuilder builds prompts for Full URD generation
type PromptBuilder struct{}

// NewPromptBuilder creates a new prompt builder
func NewPromptBuilder() *PromptBuilder {
	return &PromptBuilder{}
}

// BuildUserPrompt creates the user prompt with context and dynamic instruction
func (pb *PromptBuilder) BuildUserPrompt(ctx *ModuleContext, instructionTemplate string) string {
	var sb strings.Builder

	// Header
	sb.WriteString("Generate a COMPREHENSIVE Full URD document for the following module:\n\n")
	sb.WriteString(fmt.Sprintf("**Module Name:** %s\n", ctx.ModuleName))
	sb.WriteString(fmt.Sprintf("**Product:** %s\n", ctx.ProductName))
	if ctx.ProductVision != "" {
		sb.WriteString(fmt.Sprintf("**Vision:** %s\n", ctx.ProductVision))
	}
	sb.WriteString("\n---\n\n")

	// Context sections
	pb.writeUseCasesContext(&sb, ctx)
	pb.writeEntitiesContext(&sb, ctx)
	pb.writeBusinessRulesContext(&sb, ctx)
	pb.writeIntegrationsContext(&sb, ctx)
	pb.writePersonasContext(&sb, ctx)
	pb.writeFeaturesContext(&sb, ctx)
	pb.writeUserStoriesContext(&sb, ctx)

	// Output format specification
	sb.WriteString("\n---\n\n")
	sb.WriteString("# OUTPUT FORMAT\n\n")
	sb.WriteString("Generate a complete Full URD document following this EXACT structure:\n\n")
	sb.WriteString(instructionTemplate)

	return sb.String()
}

func (pb *PromptBuilder) writeUseCasesContext(sb *strings.Builder, ctx *ModuleContext) {
	sb.WriteString("# CONTEXT: Detailed Use Cases\n\n")
	for _, uc := range ctx.UseCases {
		sb.WriteString(fmt.Sprintf("## %s - %s\n", uc.ID, uc.Name))
		sb.WriteString(fmt.Sprintf("- **Description:** %s\n", uc.Description))
		sb.WriteString(fmt.Sprintf("- **Primary Actor:** %s\n", uc.PrimaryActor))
		sb.WriteString(fmt.Sprintf("- **Trigger:** %s\n", uc.Trigger))
		if len(uc.Preconditions) > 0 {
			sb.WriteString("- **Preconditions:**\n")
			for _, p := range uc.Preconditions {
				sb.WriteString(fmt.Sprintf("  - %s\n", p))
			}
		}
		if len(uc.Postconditions) > 0 {
			sb.WriteString("- **Postconditions:**\n")
			for _, p := range uc.Postconditions {
				sb.WriteString(fmt.Sprintf("  - %s\n", p))
			}
		}
		if len(uc.MainFlow) > 0 {
			sb.WriteString("- **Main Flow:**\n")
			for i, step := range uc.MainFlow {
				sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, step))
			}
		}
		sb.WriteString("\n")
	}
}

func (pb *PromptBuilder) writeEntitiesContext(sb *strings.Builder, ctx *ModuleContext) {
	sb.WriteString("# CONTEXT: Data Entities\n\n")
	for _, e := range ctx.Entities {
		sb.WriteString(fmt.Sprintf("## %s - %s\n", e.ID, e.Name))
		sb.WriteString(fmt.Sprintf("- **Description:** %s\n", e.Description))
		if len(e.Attributes) > 0 {
			sb.WriteString("- **Attributes:**\n")
			for _, a := range e.Attributes {
				sb.WriteString(fmt.Sprintf("  - %s\n", a))
			}
		}
		sb.WriteString("\n")
	}
}

func (pb *PromptBuilder) writeBusinessRulesContext(sb *strings.Builder, ctx *ModuleContext) {
	sb.WriteString("# CONTEXT: Business Rules\n\n")
	for _, r := range ctx.BusinessRules {
		sb.WriteString(fmt.Sprintf("- %s: %s - %s\n", r.ID, r.Name, r.Description))
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

func (pb *PromptBuilder) writePersonasContext(sb *strings.Builder, ctx *ModuleContext) {
	sb.WriteString("# CONTEXT: Personas\n\n")
	for _, p := range ctx.Personas {
		sb.WriteString(fmt.Sprintf("- %s: %s (%s)\n", p.ID, p.Name, p.Role))
	}
	sb.WriteString("\n")
}

func (pb *PromptBuilder) writeFeaturesContext(sb *strings.Builder, ctx *ModuleContext) {
	sb.WriteString("# CONTEXT: Features\n\n")
	for _, f := range ctx.Features {
		sb.WriteString(fmt.Sprintf("- %s: %s\n", f.ID, f.Name))
	}
	sb.WriteString("\n")
}

func (pb *PromptBuilder) writeUserStoriesContext(sb *strings.Builder, ctx *ModuleContext) {
	sb.WriteString("# CONTEXT: User Stories\n\n")
	for _, s := range ctx.UserStories {
		sb.WriteString(fmt.Sprintf("- %s: %s\n", s.ID, s.IWant))
	}
	sb.WriteString("\n")
}

// BuildOverviewPrompt builds the prompt for Section 1 (Mapping) & 2 (Actors) & 3 (Scope)
func (pb *PromptBuilder) BuildOverviewPrompt(ctx *ModuleContext, instructionTemplate string, previousContent string) string {
	var sb strings.Builder

	// Header
	sb.WriteString("Generate **PART 1: OVERVIEW** of the Full URD document for the following module.\n")
	sb.WriteString("You are provided with the 'URD Outline' (Previous Step). Use it as the BASE content.\n")
	sb.WriteString("You must FILL IN the tables for US Mapping, Actors, and System Boundary with detailed data, respecting the choices made in the Outline.\n\n")

	if previousContent != "" {
		sb.WriteString("--- URD OUTLINE (PREVIOUS STEP) ---\n")
		sb.WriteString(previousContent)
		sb.WriteString("\n--- END URD OUTLINE ---\n\n")
	}

	sb.WriteString(fmt.Sprintf("**Module Name:** %s\n", ctx.ModuleName))
	sb.WriteString(fmt.Sprintf("**Product:** %s\n", ctx.ProductName))
	if ctx.ProductVision != "" {
		sb.WriteString(fmt.Sprintf("**Vision:** %s\n", ctx.ProductVision))
	}
	sb.WriteString("\n---\n\n")

	// Context sections relevant to Overview
	pb.writePersonasContext(&sb, ctx)
	pb.writeUserStoriesContext(&sb, ctx)
	pb.writeFeaturesContext(&sb, ctx)

	// Output format specification
	sb.WriteString("\n---\n\n")
	sb.WriteString("# OUTPUT FORMAT\n\n")
	sb.WriteString("Generate ONLY the following sections based on the template:\n")
	sb.WriteString("1. **US -> Use Case Mapping**\n")
	sb.WriteString("2. **Actor Definition**\n")
	sb.WriteString("3. **System Boundary**\n\n")
	sb.WriteString("Start strictly with `# 1 US -> Use Case Mapping`.\n")
	sb.WriteString("Do not include the document title or introduction.\n")

	return sb.String()
}

// BuildUseCasesPrompt builds the prompt for Section 4 (Summary) & 5 (Main Flows)
func (pb *PromptBuilder) BuildUseCasesPrompt(ctx *ModuleContext, instructionTemplate string, previousContent string) string {
	var sb strings.Builder

	// Header
	sb.WriteString("Generate **PART 2: USE CASES** of the Full URD document for the following module.\n")
	sb.WriteString("You are provided with the 'URD Outline'. Use it as the BASE content.\n")
	sb.WriteString("You must generate DETAILED specifications for each Use Case identified in the Outline, including Main Flow (step-by-step), Alternative Flows, Exception Flows, Sequence Diagrams, and API Specs.\n\n")

	if previousContent != "" {
		sb.WriteString("--- URD OUTLINE (PREVIOUS STEP) ---\n")
		sb.WriteString(previousContent)
		sb.WriteString("\n--- END URD OUTLINE ---\n\n")
	}

	sb.WriteString(fmt.Sprintf("**Module Name:** %s\n", ctx.ModuleName))
	sb.WriteString("\n---\n\n")

	// Context sections relevant to Use Cases
	pb.writeUseCasesContext(&sb, ctx)
	pb.writeBusinessRulesContext(&sb, ctx)

	// Output format specification
	sb.WriteString("\n---\n\n")
	sb.WriteString("# OUTPUT FORMAT\n\n")
	sb.WriteString("Generate ONLY the following sections based on the template:\n")
	sb.WriteString("4. **Use Case Summary Table**\n")
	sb.WriteString("5. **Main Flow Sketch (High-Level)** and **Detailed Specifications** (Sections 5.x.1 to 5.x.7)\n\n")
	sb.WriteString("Start strictly with `# 4 Use Case Summary Table`.\n")

	return sb.String()
}

// BuildTechnicalPrompt builds the prompt for Section 6 (Integrations) & 7 (Data) & 8 (Technical) & 9 (Assumptions)
func (pb *PromptBuilder) BuildTechnicalPrompt(ctx *ModuleContext, instructionTemplate string, previousContent string) string {
	var sb strings.Builder

	// Header
	sb.WriteString("Generate **PART 3: TECHNICAL & DATA** of the Full URD document for the following module.\n")
	sb.WriteString("You are provided with the 'URD Outline'. Use it as the BASE content.\n")
	sb.WriteString("You must FILL IN the tables for Integrations, Data Entities, and Technical Concerns with specific technical details, respecting the Outline.\n\n")

	if previousContent != "" {
		sb.WriteString("--- URD OUTLINE (PREVIOUS STEP) ---\n")
		sb.WriteString(previousContent)
		sb.WriteString("\n--- END URD OUTLINE ---\n\n")
	}

	sb.WriteString(fmt.Sprintf("**Module Name:** %s\n", ctx.ModuleName))
	sb.WriteString("\n---\n\n")

	// Context sections relevant to Technical
	pb.writeEntitiesContext(&sb, ctx)
	pb.writeIntegrationsContext(&sb, ctx)
	pb.writeBusinessRulesContext(&sb, ctx)

	// Output format specification
	sb.WriteString("\n---\n\n")
	sb.WriteString("# OUTPUT FORMAT\n\n")
	sb.WriteString("Generate ONLY the following sections based on the template:\n")
	sb.WriteString("6. **Integration Touchpoints**\n")
	sb.WriteString("7. **Data Entities Overview**\n")
	sb.WriteString("8. **Technical Notes & Concerns**\n\n")
	sb.WriteString("Start strictly with `# 6 Integration Touchpoints`.\n")

	return sb.String()
}
