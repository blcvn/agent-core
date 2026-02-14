package registry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Tool struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  string `json:"parameters"` // JSON Schema string
	Type        string `json:"type"`       // API, CODE, MCP
}

type Client interface {
	GetTool(toolID string) (*Tool, error)
	ListTools() ([]Tool, error)
}

type client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(registryURL string) Client {
	return &client{
		baseURL: registryURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c *client) ListTools() ([]Tool, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/tools")
	if err != nil {
		return nil, fmt.Errorf("failed to call registry: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registry returned status: %d", resp.StatusCode)
	}

	var tools []Tool
	if err := json.NewDecoder(resp.Body).Decode(&tools); err != nil {
		return nil, fmt.Errorf("failed to decode tools: %w", err)
	}
	return tools, nil
}

func (c *client) GetTool(toolID string) (*Tool, error) {
	resp, err := c.httpClient.Get(fmt.Sprintf("%s/tools/%s", c.baseURL, toolID))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tool %s: %w", toolID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registry returned status: %d", resp.StatusCode)
	}

	var tool Tool
	if err := json.NewDecoder(resp.Body).Decode(&tool); err != nil {
		return nil, fmt.Errorf("failed to decode tool: %w", err)
	}
	return &tool, nil
}
