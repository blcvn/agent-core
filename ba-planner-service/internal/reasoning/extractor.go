package chat_agent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/blcvn/backend/services/ba-planner-service/internal/domain"
	v32 "github.com/blcvn/backend/services/ba-planner-service/internal/domain/v3.2"
)

// SectionExtractor identifies affected sections in the document
type SectionExtractor struct {
	llmService domain.LLMService
}

// NewSectionExtractor creates a new SectionExtractor
func NewSectionExtractor(llmService domain.LLMService) *SectionExtractor {
	return &SectionExtractor{
		llmService: llmService,
	}
}

// ExtractAffectedSections identifies sections affected by the comment
func (e *SectionExtractor) ExtractAffectedSections(ctx context.Context, comment string, doc *v32.Document, kg *v32.RequirementGraph) ([]AffectedSection, error) {
	// 1. Build Prompt
	systemPrompt := `You are an expert at analyzing comments on requirement documents to identify affected sections.
Your task is to identify which specific sections of the document are affected by a comment.

Return ONLY a JSON array of affected sections:
[
  {
    "section_type": "use_case|actor|entity|integration|business_rule|flow_step|general",
    "section_id": "UC-001|ACT-001|ENT-001|etc.",
    "section_name": "human-readable name",
    "confidence_score": 0.9,
    "reasoning": "why this section is affected"
  }
]

# DETECTION RULES:
1. Explicit IDs: "UC-001", "Step 3"
2. Implicit References: "login flow" -> Login Use Case
3. Keywords: "add user" -> User Actor/Entity
4. Context: "this use case" -> most recent/relevant
`

	// Format document sections for context
	docContext := formatDocumentContext(doc)

	userPrompt := fmt.Sprintf("# DOCUMENT TYPE\n%s\n\n# COMMENT\n%s\n\n# AVAILABLE SECTIONS\n%s\n\nIdentify affected sections.", doc.Tier, comment, docContext)

	// 2. Call LLM
	response, err := e.llmService.Chat(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to call LLM for extraction: %w", err)
	}

	// 3. Parse Response
	content := cleanJSON(response)
	var sections []AffectedSection
	if err := json.Unmarshal([]byte(content), &sections); err != nil {
		return nil, fmt.Errorf("failed to parse sections JSON: %w", err)
	}

	// 4. Enrich with Current Content & KG Nodes
	for i := range sections {
		// Find related nodes in KG
		sections[i].RelatedNodeIDs = findRelatedNodesInKG(kg, sections[i].SectionType, sections[i].SectionID)

		// Extract current content
		sections[i].CurrentContent = extractCurrentContent(doc, sections[i].SectionType, sections[i].SectionID)
	}

	return sections, nil
}

func formatDocumentContext(doc *v32.Document) string {
	// TODO: Implement proper formatting based on doc structure (Index/Outline/Full)
	// For now return a placeholder or simplified list
	return "TODO: List use cases, actors, etc. from doc content"
}

func findRelatedNodesInKG(kg *v32.RequirementGraph, sectionType, sectionID string) []string {
	var related []string
	if kg == nil {
		return related
	}
	// Simple lookup for now
	for _, node := range kg.Nodes {
		// Assuming node.Metadata["id"] matches sectionID
		if id, ok := node.Metadata["id"].(string); ok && id == sectionID {
			related = append(related, node.ID)
			// Add connected nodes...
		} else if node.ReferenceID == sectionID {
			// Also check ReferenceID (e.g. US-001)
			related = append(related, node.ID)
		}
	}
	return related
}

func extractCurrentContent(doc *v32.Document, sectionType, sectionID string) string {
	// TODO: Parse doc.Content or structured fields to find specific content
	return ""
}
