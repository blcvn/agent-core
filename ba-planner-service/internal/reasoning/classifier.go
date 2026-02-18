package chat_agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/blcvn/backend/services/ba-agent-service/domain"
)

// IntentClassifier is responsible for classifying the intent of a user's comment
type IntentClassifier struct {
	llmService domain.LLMService
}

// NewIntentClassifier creates a new IntentClassifier
func NewIntentClassifier(llmService domain.LLMService) *IntentClassifier {
	return &IntentClassifier{
		llmService: llmService,
	}
}

// ClassifyCommentIntent classifies the intent of a comment
func (c *IntentClassifier) ClassifyCommentIntent(ctx context.Context, comment string) (*CommentIntent, error) {
	systemPrompt := `You are an expert at analyzing comments on requirement documents to classify their intent.
Your task is to determine the type, scope, and urgency of the comment, and extract keywords.

Return ONLY a JSON object:
{
  "type": "add|modify|delete|question|clarification",
  "scope": "specific_section|entire_document|cross_cutting",
  "urgency": "low|medium|high",
  "keywords": ["keyword1", "keyword2"]
}

# CLASSIFICATION RULES:

1. **Type:**
   - "add": Request to add new content (use case, actor, etc.)
   - "modify": Request to change existing content
   - "delete": Request to remove content
   - "question": Asking for information without requesting a change
   - "clarification": Pointing out ambiguity

2. **Scope:**
   - "specific_section": Targets a specific ID or named section
   - "entire_document": General feedback (e.g., "rewrite the whole doc")
   - "cross_cutting": Affects multiple areas (e.g., "rename User to Customer everywhere")

3. **Urgency:**
   - "high": Critical issues, blockers, security flaws
   - "medium": Routine changes, corrections
   - "low": Suggestions, minor edits
`

	userPrompt := fmt.Sprintf("# COMMENT\n%s\n\nClassify this comment.", comment)

	response, err := c.llmService.Chat(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to call LLM for classification: %w", err)
	}

	content := cleanJSON(response)
	var intent CommentIntent
	if err := json.Unmarshal([]byte(content), &intent); err != nil {
		return nil, fmt.Errorf("failed to parse intent JSON: %w", err)
	}

	return &intent, nil
}

func cleanJSON(text string) string {
	text = strings.TrimSpace(text)
	if strings.HasPrefix(text, "```json") {
		text = strings.TrimPrefix(text, "```json")
		text = strings.TrimSuffix(text, "```")
	} else if strings.HasPrefix(text, "```") {
		text = strings.TrimPrefix(text, "```")
		text = strings.TrimSuffix(text, "```")
	}
	return strings.TrimSpace(text)
}
