package interaction

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/blcvn/backend/services/pkg/domain/tool"
)

// AskUserTool allows the agent to ask the user for input
type AskUserTool struct{}

// NewAskUserTool creates a new AskUserTool
func NewAskUserTool() *AskUserTool {
	return &AskUserTool{}
}

func (t *AskUserTool) Name() string {
	return "ask_user"
}

func (t *AskUserTool) Description() string {
	return "Ask the user for information or confirmation. Use this when you need more details to proceed."
}

func (t *AskUserTool) Schema() string {
	return `{
		"type": "object",
		"properties": {
			"question": {
				"type": "string",
				"description": "The question to ask the user"
			}
		},
		"required": ["question"]
	}`
}

type AskUserInput struct {
	Question string `json:"question"`
}

func (t *AskUserTool) Execute(ctx context.Context, input string) (string, error) {
	var params AskUserInput
	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("invalid input: %v", err)
	}
	return "", &tool.ErrRequiresInput{Question: params.Question}
}
