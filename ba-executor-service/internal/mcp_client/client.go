package mcp_client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sync"

	"github.com/blcvn/backend/services/pkg/mcp"
)

type Client struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Scanner
	id     int
	mu     sync.Mutex
}

func NewClient(command string, args ...string) (*Client, error) {
	cmd := exec.Command(command, args...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return &Client{
		cmd:    cmd,
		stdin:  stdin,
		stdout: bufio.NewScanner(stdoutPipe),
		id:     0,
	}, nil
}

func (c *Client) ListTools() (*mcp.ListToolsResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.id++
	req := mcp.Request{
		JsonRPC: "2.0",
		Method:  "tools/list",
		ID:      c.id,
	}

	if err := c.sendRequest(req); err != nil {
		return nil, err
	}

	resp, err := c.readResponse()
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("mcp error: %s", resp.Error.Message)
	}

	var result mcp.ListToolsResult
	bytes, _ := json.Marshal(resp.Result)
	if err := json.Unmarshal(bytes, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) CallTool(name string, args map[string]interface{}) (*mcp.CallToolResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.id++
	paramsBytes, _ := json.Marshal(mcp.CallToolParams{
		Name:      name,
		Arguments: args,
	})

	req := mcp.Request{
		JsonRPC: "2.0",
		Method:  "tools/call",
		Params:  json.RawMessage(paramsBytes),
		ID:      c.id,
	}

	if err := c.sendRequest(req); err != nil {
		return nil, err
	}

	resp, err := c.readResponse()
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("mcp error: %s", resp.Error.Message)
	}

	var result mcp.CallToolResult
	bytes, _ := json.Marshal(resp.Result)
	if err := json.Unmarshal(bytes, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) sendRequest(req mcp.Request) error {
	bytes, err := json.Marshal(req)
	if err != nil {
		return err
	}
	_, err = c.stdin.Write(append(bytes, '\n'))
	return err
}

func (c *Client) readResponse() (*mcp.Response, error) {
	if c.stdout.Scan() {
		line := c.stdout.Bytes()
		var resp mcp.Response
		if err := json.Unmarshal(line, &resp); err != nil {
			return nil, err
		}
		return &resp, nil
	}
	return nil, c.stdout.Err()
}

func (c *Client) Close() {
	c.stdin.Close()
	c.cmd.Wait()
}
