package analyzer

import (
	"context"
	"encoding/json"
	"fmt"

	aiproxy "github.com/blcvn/kratos-proto/go/ai-proxy"
)

// FeatureAnalyzerTool analyzes feature requirements
type FeatureAnalyzerTool struct {
	aiProxyClient aiproxy.AIProxyServiceClient
}

// NewFeatureAnalyzerTool creates a new feature analyzer tool
func NewFeatureAnalyzerTool(aiProxyClient aiproxy.AIProxyServiceClient) *FeatureAnalyzerTool {
	return &FeatureAnalyzerTool{
		aiProxyClient: aiProxyClient,
	}
}

func (t *FeatureAnalyzerTool) Name() string {
	return "FeatureAnalyzer"
}

func (t *FeatureAnalyzerTool) Description() string {
	return "Analyze feature requirements and identify gaps, dependencies, risks. Input: {\"feature\": \"feature description\"}"
}

func (t *FeatureAnalyzerTool) Schema() string {
	return `{
		"type": "object",
		"properties": {
			"feature": {"type": "string", "description": "Feature description to analyze"}
		},
		"required": ["feature"]
	}`
}

func (t *FeatureAnalyzerTool) Execute(ctx context.Context, input string) (string, error) {
	var params struct {
		Feature string `json:"feature"`
	}

	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	prompt := fmt.Sprintf(`Analyze the following feature requirement and provide:

1. **Functional Breakdown**: List all functional components
2. **Dependencies**: Identify technical and business dependencies
3. **Risks**: Highlight potential risks and mitigation strategies
4. **Gaps**: Identify missing information or unclear requirements
5. **Acceptance Criteria**: Suggest acceptance criteria

Feature:
%s

Provide a structured analysis in JSON format.`, params.Feature)

	resp, err := t.aiProxyClient.Complete(ctx, &aiproxy.CompleteRequest{
		Payload: &aiproxy.CompletePayload{
			ModelId:     "claude-3-5-sonnet-20241022",
			Prompt:      prompt,
			Temperature: 0.4,
			MaxTokens:   4096,
		},
	})

	if err != nil {
		return "", fmt.Errorf("failed to analyze feature: %w", err)
	}

	if resp.Result.Code != aiproxy.ResultCode_SUCCESS {
		return "", fmt.Errorf("AI Proxy error: %s", resp.Result.Message)
	}

	return resp.Completion.Text, nil
}

// ValidationTool validates document completeness
type ValidationTool struct {
	aiProxyClient aiproxy.AIProxyServiceClient
}

// NewValidationTool creates a new validation tool
func NewValidationTool(aiProxyClient aiproxy.AIProxyServiceClient) *ValidationTool {
	return &ValidationTool{
		aiProxyClient: aiProxyClient,
	}
}

func (t *ValidationTool) Name() string {
	return "ValidationTool"
}

func (t *ValidationTool) Description() string {
	return "Validate document completeness and quality. Input: {\"document\": \"document content\", \"type\": \"BRD|SRS|PRD\"}"
}

func (t *ValidationTool) Schema() string {
	return `{
		"type": "object",
		"properties": {
			"document": {"type": "string", "description": "Document content to validate"},
			"type": {"type": "string", "enum": ["BRD", "SRS", "PRD"], "description": "Document type"}
		},
		"required": ["document", "type"]
	}`
}

func (t *ValidationTool) Execute(ctx context.Context, input string) (string, error) {
	var params struct {
		Document string `json:"document"`
		Type     string `json:"type"`
	}

	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	prompt := fmt.Sprintf(`Validate the following %s document for completeness and quality:

Document:
%s

Check for:
1. All required sections present
2. Clarity and specificity of requirements
3. Consistency across sections
4. Missing information or gaps
5. Ambiguous language

Provide a validation report with:
- Overall score (0-100)
- Issues found (categorized by severity: Critical, Major, Minor)
- Recommendations for improvement

Format as JSON.`, params.Type, params.Document)

	resp, err := t.aiProxyClient.Complete(ctx, &aiproxy.CompleteRequest{
		Payload: &aiproxy.CompletePayload{
			ModelId:     "claude-3-5-sonnet-20241022",
			Prompt:      prompt,
			Temperature: 0.3,
			MaxTokens:   4096,
		},
	})

	if err != nil {
		return "", fmt.Errorf("failed to validate document: %w", err)
	}

	if resp.Result.Code != aiproxy.ResultCode_SUCCESS {
		return "", fmt.Errorf("AI Proxy error: %s", resp.Result.Message)
	}

	return resp.Completion.Text, nil
}
