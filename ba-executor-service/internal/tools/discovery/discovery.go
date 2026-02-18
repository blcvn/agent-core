package discovery

import (
	"context"
	"encoding/json"
	"fmt"

	aiproxy "github.com/blcvn/kratos-proto/go/ai-proxy"
	prompt "github.com/blcvn/kratos-proto/go/prompt"
)

// DiscoveryTool helps in gathering requirements
type DiscoveryTool struct {
	aiProxyClient aiproxy.AIProxyServiceClient
	promptClient  prompt.PromptServiceClient
}

// NewDiscoveryTool creates a new discovery tool
func NewDiscoveryTool(
	aiProxyClient aiproxy.AIProxyServiceClient,
	promptClient prompt.PromptServiceClient,
) *DiscoveryTool {
	return &DiscoveryTool{
		aiProxyClient: aiProxyClient,
		promptClient:  promptClient,
	}
}

func (t *DiscoveryTool) Name() string {
	return "DiscoveryTool"
}

func (t *DiscoveryTool) Description() string {
	return "Gather project requirements. Input: {\"action\": \"start|continue\", \"projectName\": \"...\", \"userMessage\": \"...\", \"history\": [...]}"
}

func (t *DiscoveryTool) Schema() string {
	return `{
		"type": "object",
		"properties": {
			"action": {"type": "string", "enum": ["start", "continue"], "description": "Action to perform"},
			"projectName": {"type": "string", "description": "Name of the project"},
			"projectDescription": {"type": "string", "description": "Initial description"},
			"userMessage": {"type": "string", "description": "Latest user response"},
			"history": {"type": "array", "description": "Conversation history context"}
		},
		"required": ["action"]
	}`
}

func (t *DiscoveryTool) Execute(ctx context.Context, input string) (string, error) {
	var params struct {
		Action             string          `json:"action"`
		ProjectName        string          `json:"projectName"`
		ProjectDescription string          `json:"projectDescription"`
		UserMessage        string          `json:"userMessage"`
		History            json.RawMessage `json:"history"`
	}

	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	var promptText string
	// Logic to determine prompt based on action
	if params.Action == "start" {
		promptText = t.getStartPrompt(params.ProjectName, params.ProjectDescription)
	} else if params.Action == "continue" {
		promptText = t.getContinuePrompt(params.UserMessage, string(params.History))
	} else {
		return "", fmt.Errorf("unknown action: %s", params.Action)
	}

	// Call AI Proxy
	resp, err := t.aiProxyClient.Complete(ctx, &aiproxy.CompleteRequest{
		Payload: &aiproxy.CompletePayload{
			ModelId:     "claude-3-5-sonnet-20241022", // Use config model ID ideally
			Prompt:      promptText,
			Temperature: 0.7,
			MaxTokens:   4096,
		},
	})

	if err != nil {
		return "", fmt.Errorf("failed to execute discovery: %w", err)
	}

	if resp.Result.Code != aiproxy.ResultCode_SUCCESS {
		return "", fmt.Errorf("AI Proxy error: %s", resp.Result.Message)
	}

	// Parse response to ensure JSON format? The tool should return the AI response directly,
	// checking if it's valid JSON might be good but let's trust the prompt.
	return resp.Completion.Text, nil
}

func (t *DiscoveryTool) getStartPrompt(projectName, description string) string {
	// TODO: Fetch from Prompt Service if available
	return fmt.Sprintf(`Bắt đầu phân tích yêu cầu cho dự án: "%s"
%s

Hãy bắt đầu giai đoạn THU THẬP THÔNG TIN.
Hỏi câu hỏi ĐẦU TIÊN về CÁC BÊN LIÊN QUAN.
Nhớ: CHỈ HỎI MỘT CÂU DUY NHẤT. Trả lời bằng tiếng Việt.

BẠN PHẢI LUÔN TRẢ VỀ RESPONSE THEO FORMAT JSON SAU:
{
  "message": "Nội dung giới thiệu và tóm tắt ngán gọn",
  "questions": [
    {
      "id": "q1",
      "category": "Stakeholders",
      "text": "Câu hỏi đầy đủ ở đây?",
      "multiSelect": true,
      "options": ["Option A", "Option B", "Option C"]
    }
  ]
}`, projectName, description)
}

func (t *DiscoveryTool) getContinuePrompt(userMessage, history string) string {
	return fmt.Sprintf(`User vừa trả lời: "%s"

Dựa trên câu trả lời này và lịch sử:
%s

Hãy:
1. Tóm tắt và confirm thông tin
2. Tạo câu hỏi tiếp theo phù hợp với ngữ cảnh

BẠN PHẢI LUÔN TRẢ VỀ RESPONSE THEO FORMAT JSON SAU:
{
  "message": "Tóm tắt câu trả lời + transition",
  "questions": [
    {
      "id": "qX",
      "category": "...",
      "text": "Câu hỏi?",
      "multiSelect": true/false,
      "options": [...]
    }
  ]
}`, userMessage, history)
}
