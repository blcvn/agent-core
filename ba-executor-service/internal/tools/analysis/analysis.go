package analysis

import (
	"context"
	"encoding/json"
	"fmt"

	aiproxy "github.com/blcvn/kratos-proto/go/ai-proxy"
	prompt "github.com/blcvn/kratos-proto/go/prompt"
)

// AnalysisTool extracts requirements from discovery data
type AnalysisTool struct {
	aiProxyClient aiproxy.AIProxyServiceClient
	promptClient  prompt.PromptServiceClient
}

// NewAnalysisTool creates a new analysis tool
func NewAnalysisTool(
	aiProxyClient aiproxy.AIProxyServiceClient,
	promptClient prompt.PromptServiceClient,
) *AnalysisTool {
	return &AnalysisTool{
		aiProxyClient: aiProxyClient,
		promptClient:  promptClient,
	}
}

func (t *AnalysisTool) Name() string {
	return "AnalysisTool"
}

func (t *AnalysisTool) Description() string {
	return "Extract requirements (FR, NFR, BR, DR) from discovery data. Input: {\"action\": " + "\"extract\"" + ", \"discoveryData\": {...}}"
}

func (t *AnalysisTool) Schema() string {
	return `{
		"type": "object",
		"properties": {
			"action": {"type": "string", "enum": ["extract"], "description": "Action to perform"},
			"discoveryData": {"type": "object", "description": "Collected discovery data"}
		},
		"required": ["action", "discoveryData"]
	}`
}

func (t *AnalysisTool) Execute(ctx context.Context, input string) (string, error) {
	var params struct {
		Action        string      `json:"action"`
		DiscoveryData interface{} `json:"discoveryData"`
	}

	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	var promptText string
	if params.Action == "extract" {
		discoveryJSON, _ := json.MarshalIndent(params.DiscoveryData, "", "  ")
		promptText = t.getExtractPrompt(string(discoveryJSON))
	} else {
		return "", fmt.Errorf("unknown action: %s", params.Action)
	}

	// Call AI Proxy
	resp, err := t.aiProxyClient.Complete(ctx, &aiproxy.CompleteRequest{
		Payload: &aiproxy.CompletePayload{
			ModelId:     "claude-3-5-sonnet-20241022",
			Prompt:      promptText,
			Temperature: 0.3,
			MaxTokens:   4096,
		},
	})

	if err != nil {
		return "", fmt.Errorf("failed to execute analysis: %w", err)
	}

	if resp.Result.Code != aiproxy.ResultCode_SUCCESS {
		return "", fmt.Errorf("AI Proxy error: %s", resp.Result.Message)
	}

	return resp.Completion.Text, nil
}

func (t *AnalysisTool) getExtractPrompt(discoveryData string) string {
	// TODO: Fetch from Prompt Service
	return fmt.Sprintf(`Phân tích dữ liệu thu thập sau và trích xuất TẤT CẢ yêu cầu:

%s

Hãy tạo danh sách yêu cầu với format (viết hoàn toàn bằng tiếng Việt):

## YÊU CẦU CHỨC NĂNG
| ID | Tiêu đề | Mô tả | Tiêu chí chấp nhận | Độ ưu tiên | Story Points |
|---|---|---|---|---|---|
| FR-001 | ... | ... | ... | Bắt buộc | 5 |

## YÊU CẦU PHI CHỨC NĂNG
| ID | Tiêu đề | Mô tả | Loại | Chỉ số đo lường | Độ ưu tiên |
|---|---|---|---|---|---|
| NFR-001 | ... | ... | Hiệu năng | P90 < 3s | Bắt buộc |

## QUY TẮC NGHIỆP VỤ
| ID | Tiêu đề | Điều kiện | Hành động | Lý do |
|---|---|---|---|---|
| BR-001 | ... | NẾU ... | THÌ ... | ... |

## YÊU CẦU DỮ LIỆU
| ID | Tiêu đề | Thực thể | Thời gian lưu trữ | Bảo mật |
|---|---|---|---|---|
| DR-001 | ... | Người dùng, Hồ sơ | 7 năm | Dữ liệu cá nhân |

Hãy trích xuất NHIỀU yêu cầu (ít nhất 5 FR, 3 NFR, 2 BR, 2 DR). Viết toàn bộ bằng tiếng Việt.`, discoveryData)
}
