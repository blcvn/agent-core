package chat_agent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/blcvn/ba-shared-libs/pkg/domain"
	v32 "github.com/blcvn/ba-shared-libs/pkg/domain/v3.2"
)

// ModificationGenerator generates document modifications using LLM
type ModificationGenerator struct {
	llmService domain.LLMService
}

// NewModificationGenerator creates a new ModificationGenerator
func NewModificationGenerator(llmService domain.LLMService) *ModificationGenerator {
	return &ModificationGenerator{
		llmService: llmService,
	}
}

// GenerateModifications generates modifications based on context and intent
func (g *ModificationGenerator) GenerateModifications(ctx context.Context, comment string, intent *CommentIntent, mCtx *ModificationContext, docTier v32.RequirementTier) ([]Modification, error) {
	// 1. Build System Prompt
	systemPrompt := `You are a Senior Business Analyst expert at modifying requirement documents based on stakeholder feedback.
Your task is to generate specific modifications to document sections based on a comment.

CRITICAL RULES:
1. Only modify what the comment asks for.
2. Maintain document structure and format.
3. Ensure consistency with related items.
4. Preserve traceability.
5. Follow document type conventions.

OUTPUT FORMAT:
Return a JSON array of modifications:
[
  {
    "modification_id": "MOD-001",
    "section_type": "use_case|actor|entity|integration|business_rule|flow_step",
    "section_id": "UC-001|ACT-001|etc.",
    "action_type": "add|modify|delete|no_change",
    "old_content": "current content (for modify/delete)",
    "new_content": "updated content (for add/modify)",
    "reasoning": "why this change is being made",
    "impact_analysis": "what other sections might be affected"
  }
]
`

	// 2. Build User Prompt
	userPrompt := fmt.Sprintf(`
# COMMENT
%s

# INTENT
Type: %s
Scope: %s
Urgency: %s

# DOCUMENT TYPE
%s

# AFFECTED SECTIONS
%s

# RELATED CONTEXT
User Stories: %d
Features: %d
Business Rules: %d

Generate specific modifications.
`, comment, intent.Type, intent.Scope, intent.Urgency, docTier, formatAffectedSections(mCtx.AffectedSections), len(mCtx.RelatedUserStories), len(mCtx.RelatedFeatures), len(mCtx.RelatedBusinessRules))

	// 3. Call LLM
	response, err := g.llmService.Chat(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to call LLM for modification generation: %w", err)
	}

	// 4. Parse Response
	content := cleanJSON(response)
	var modifications []Modification
	if err := json.Unmarshal([]byte(content), &modifications); err != nil {
		return nil, fmt.Errorf("failed to parse modifications JSON: %w", err)
	}

	// 5. Add IDs if missing
	for i := range modifications {
		if modifications[i].ModificationID == "" {
			modifications[i].ModificationID = fmt.Sprintf("MOD-%d", time.Now().UnixNano())
		}
	}

	return modifications, nil
}

func formatAffectedSections(sections []AffectedSection) string {
	// TODO: Format properly
	return fmt.Sprintf("%d sections affected", len(sections))
}
