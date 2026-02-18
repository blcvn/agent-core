package document

import (
	"context"
	"encoding/json"
	"fmt"

	aiproxy "github.com/blcvn/kratos-proto/go/ai-proxy"
	prompt "github.com/blcvn/kratos-proto/go/prompt"
)

// DocumentGeneratorTool generates BA documents (BRD, SRS, etc.)
type DocumentGeneratorTool struct {
	aiProxyClient aiproxy.AIProxyServiceClient
	promptClient  prompt.PromptServiceClient
}

// NewDocumentGeneratorTool creates a new document generator tool
func NewDocumentGeneratorTool(
	aiProxyClient aiproxy.AIProxyServiceClient,
	promptClient prompt.PromptServiceClient,
) *DocumentGeneratorTool {
	return &DocumentGeneratorTool{
		aiProxyClient: aiProxyClient,
		promptClient:  promptClient,
	}
}

func (t *DocumentGeneratorTool) Name() string {
	return "DocumentGenerator"
}

func (t *DocumentGeneratorTool) Description() string {
	return "Generate BA documents (BRD, SRS, PRD). Input: {\"type\": \"BRD|SRS|PRD\", \"requirements\": \"requirements text\", \"context\": \"additional context\"}"
}

func (t *DocumentGeneratorTool) Schema() string {
	return `{
		"type": "object",
		"properties": {
			"type": {"type": "string", "enum": ["BRD", "SRS", "PRD"], "description": "Document type"},
			"requirements": {"type": "string", "description": "Requirements to document"},
			"context": {"type": "string", "description": "Additional context"}
		},
		"required": ["type", "requirements"]
	}`
}

func (t *DocumentGeneratorTool) Execute(ctx context.Context, input string) (string, error) {
	var params struct {
		Type         string `json:"type"`
		Requirements string `json:"requirements"`
		Context      string `json:"context"`
	}

	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	// Try to get template from Prompt Service
	templateName := fmt.Sprintf("ba-%s-template", params.Type)
	var promptText string

	if t.promptClient != nil {
		// First, list templates to find by name
		listResp, err := t.promptClient.ListTemplates(ctx, &prompt.ListTemplatesRequest{
			Metadata:    &prompt.Metadata{},
			Signature:   &prompt.Signature{},
			Environment: prompt.Environment_PROD,
		})

		var templateID string
		if err == nil && listResp.Result.Code == prompt.ResultCode_SUCCESS {
			// Find template by name
			for _, tmpl := range listResp.Templates {
				if tmpl.Name == templateName {
					templateID = tmpl.Id
					break
				}
			}
		}

		// If template found, render it
		if templateID != "" {
			renderResp, err := t.promptClient.RenderTemplate(ctx, &prompt.RenderTemplateRequest{
				Metadata:  &prompt.Metadata{},
				Signature: &prompt.Signature{},
				Payload: &prompt.RenderTemplatePayload{
					TemplateId: templateID,
					Variables: map[string]string{
						"requirements": params.Requirements,
						"context":      params.Context,
						"type":         params.Type,
					},
				},
			})

			if err == nil && renderResp.Result.Code == prompt.ResultCode_SUCCESS {
				promptText = renderResp.Rendered.RenderedText
			}
		}
	}

	// Fallback to default prompt if template not found or service unavailable
	if promptText == "" {
		promptText = t.getDefaultPrompt(params.Type, params.Requirements, params.Context)
	}

	// Generate document using AI Proxy
	resp, err := t.aiProxyClient.Complete(ctx, &aiproxy.CompleteRequest{
		Payload: &aiproxy.CompletePayload{
			ModelId:     "claude-3-5-sonnet-20241022",
			Prompt:      promptText,
			Temperature: 0.5,
			MaxTokens:   8192,
		},
	})

	if err != nil {
		return "", fmt.Errorf("failed to generate document: %w", err)
	}

	if resp.Result.Code != aiproxy.ResultCode_SUCCESS {
		return "", fmt.Errorf("AI Proxy error: %s", resp.Result.Message)
	}

	return resp.Completion.Text, nil
}

func (t *DocumentGeneratorTool) getDefaultPrompt(docType, requirements, context string) string {
	contextStr := ""
	if context != "" {
		contextStr = fmt.Sprintf("\n\nAdditional Context:\n%s", context)
	}

	return fmt.Sprintf(`Generate a comprehensive %s (Business Requirements Document) based on the following requirements:

Requirements:
%s%s

Please structure the document with:
1. Executive Summary
2. Business Objectives
3. Scope
4. Functional Requirements
5. Non-Functional Requirements
6. Assumptions and Constraints
7. Success Criteria

Use clear, professional language and include all necessary details.`, docType, requirements, contextStr)
}
