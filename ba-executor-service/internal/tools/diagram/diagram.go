package diagram

import (
	"context"
	"encoding/json"
	"fmt"

	aiproxy "github.com/blcvn/kratos-proto/go/ai-proxy"
)

// DiagramGeneratorTool generates Mermaid diagrams from requirements
type DiagramGeneratorTool struct {
	aiProxyClient aiproxy.AIProxyServiceClient
}

// NewDiagramGeneratorTool creates a new diagram generator tool
func NewDiagramGeneratorTool(aiProxyClient aiproxy.AIProxyServiceClient) *DiagramGeneratorTool {
	return &DiagramGeneratorTool{
		aiProxyClient: aiProxyClient,
	}
}

func (t *DiagramGeneratorTool) Name() string {
	return "DiagramGenerator"
}

func (t *DiagramGeneratorTool) Description() string {
	return "Generate Mermaid diagrams from requirements. Input: {\"type\": \"flowchart|sequence|class|er\", \"description\": \"what to diagram\"}"
}

func (t *DiagramGeneratorTool) Schema() string {
	return `{
		"type": "object",
		"properties": {
			"type": {"type": "string", "enum": ["flowchart", "sequence", "class", "er"], "description": "Diagram type"},
			"description": {"type": "string", "description": "What to diagram"}
		},
		"required": ["type", "description"]
	}`
}

func (t *DiagramGeneratorTool) Execute(ctx context.Context, input string) (string, error) {
	var params struct {
		Type        string `json:"type"`
		Description string `json:"description"`
	}

	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	// Build prompt for diagram generation
	prompt := fmt.Sprintf(`Generate a Mermaid %s diagram for the following:

%s

Return ONLY the Mermaid code, no explanations. Start with the diagram type declaration.`, params.Type, params.Description)

	// Call AI Proxy to generate diagram
	resp, err := t.aiProxyClient.Complete(ctx, &aiproxy.CompleteRequest{
		Payload: &aiproxy.CompletePayload{
			ModelId:     "claude-3-5-sonnet-20241022",
			Prompt:      prompt,
			Temperature: 0.3,
			MaxTokens:   2048,
		},
	})

	if err != nil {
		return "", fmt.Errorf("failed to generate diagram: %w", err)
	}

	if resp.Result.Code != aiproxy.ResultCode_SUCCESS {
		return "", fmt.Errorf("AI Proxy error: %s", resp.Result.Message)
	}

	return resp.Completion.Text, nil
}
