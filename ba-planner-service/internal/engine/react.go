package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	executorpb "github.com/blcvn/backend/services/proto/executor"
	aiproxy "github.com/blcvn/kratos-proto/go/ai-proxy"
	baagent "github.com/blcvn/kratos-proto/go/ba-agent"
)

// ReActEngine implements the Reasoning + Acting loop
type ReActEngine struct {
	aiProxyClient  aiproxy.AIProxyServiceClient
	executorClient executorpb.ExecutorServiceClient
	modelID        string
	maxIterations  int
}

// NewReActEngine creates a new ReAct engine
func NewReActEngine(
	aiProxyClient aiproxy.AIProxyServiceClient,
	executorClient executorpb.ExecutorServiceClient,
	modelID string,
	maxIterations int,
) *ReActEngine {
	return &ReActEngine{
		aiProxyClient:  aiProxyClient,
		executorClient: executorClient,
		modelID:        modelID,
		maxIterations:  maxIterations,
	}
}

// ReActStep represents one iteration of the ReAct loop
type ReActStep struct {
	Thought     string
	Action      string
	ActionInput string
	Observation string
	TimestampMs int64
	IsFinal     bool
}

// ExecuteResult contains the final result of task execution
type ExecuteResult struct {
	FinalAnswer     string
	Steps           []*ReActStep
	TotalIterations int
	ToolCalls       int
	LLMCalls        int
	DurationMs      int64
	TotalCost       float32
	TotalTokens     int
	Status          string // "COMPLETED", "WAITING_FOR_INPUT", "FAILED"
	CurrentQuestion string // If WAITING_FOR_INPUT
}

// Execute runs the ReAct loop for a given task
func (e *ReActEngine) Execute(ctx context.Context, task string, payload *baagent.ExecuteTaskPayload, onStep func(*ReActStep)) (*ExecuteResult, error) {
	// TODO: Extract context from payload if/when added to proto
	contextMap := map[string]string{}

	startTime := time.Now()
	scratchpad := []string{}
	steps := []*ReActStep{}
	llmCalls := 0
	toolCalls := 0
	totalCost := float32(0.0)
	totalTokens := 0
	startIteration := 0

	// Hydrate state from payload history if available
	if len(payload.History) > 0 {
		for _, hStep := range payload.History {
			// Reconstruct step
			step := &ReActStep{
				Thought:     hStep.Thought,
				Action:      hStep.Action,
				ActionInput: hStep.GetActionParams()["input"], // Map retrieval
				Observation: hStep.Observation,
				TimestampMs: 0, // Not persisted, irrelevant for prompting
				IsFinal:     false,
			}
			steps = append(steps, step)
			// Reconstruct scratchpad
			scratchpad = append(scratchpad,
				fmt.Sprintf("Thought: %s", step.Thought),
				fmt.Sprintf("Action: %s", step.Action),
				fmt.Sprintf("Action Input: %s", step.ActionInput),
				fmt.Sprintf("Observation: %s", step.Observation),
			)
		}
		startIteration = len(steps)
	}

	// Build system prompt with tool descriptions
	systemPrompt, err := e.buildSystemPrompt(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build system prompt: %w", err)
	}

	for i := startIteration; i < e.maxIterations; i++ {
		// Build prompt with scratchpad
		prompt := e.buildPrompt(task, contextMap, scratchpad, systemPrompt)

		// THOUGHT: Call LLM to reason about next action
		llmResponse, err := e.callLLM(ctx, prompt)
		if err != nil {
			return nil, fmt.Errorf("LLM call failed: %w", err)
		}

		llmCalls++
		totalCost += llmResponse.Cost
		totalTokens += llmResponse.TokensUsed

		// Parse the response
		thought, action, actionInput, isFinal := e.parseResponse(llmResponse.Text)

		step := &ReActStep{
			Thought:     thought,
			Action:      action,
			ActionInput: actionInput,
			TimestampMs: time.Now().UnixMilli(),
			IsFinal:     isFinal,
		}

		// Check if task is complete
		if isFinal {
			step.Observation = "Task completed"
			steps = append(steps, step)
			if onStep != nil {
				onStep(step)
			}

			return &ExecuteResult{
				FinalAnswer:     e.extractFinalAnswer(llmResponse.Text),
				Steps:           steps,
				TotalIterations: i + 1,
				ToolCalls:       toolCalls,
				LLMCalls:        llmCalls,
				DurationMs:      time.Since(startTime).Milliseconds(),
				TotalCost:       totalCost,
				TotalTokens:     totalTokens,
				Status:          "COMPLETED",
			}, nil
		}

		// ACTION: Execute the tool via Executor Service
		// TODO: Get SessionID from context or payload
		sessionID := "default-session"

		toolResp, err := e.executorClient.ExecuteTool(ctx, &executorpb.ExecuteToolRequest{
			ToolName:  action,
			InputJson: actionInput,
			SessionId: sessionID,
		})

		var observation string
		if err != nil {
			observation = fmt.Sprintf("Error executing tool (gRPC): %v", err)
		} else if !toolResp.Success {
			observation = fmt.Sprintf("Error executing tool: %s", toolResp.Error)
			// Check for interruption request
			if strings.Contains(toolResp.Error, "requires input") {
				step.Observation = "Waiting for user input..."
				steps = append(steps, step)
				if onStep != nil {
					onStep(step)
				}
				return &ExecuteResult{
					FinalAnswer:     "",
					Steps:           steps,
					TotalIterations: i,
					Status:          "WAITING_FOR_INPUT",
					CurrentQuestion: toolResp.Error,
					ToolCalls:       toolCalls,
					LLMCalls:        llmCalls,
					DurationMs:      time.Since(startTime).Milliseconds(),
					TotalCost:       totalCost,
					TotalTokens:     totalTokens,
				}, nil
			}
		} else {
			observation = toolResp.Output
			toolCalls++
		}

		step.Observation = observation
		steps = append(steps, step)
		if onStep != nil {
			onStep(step)
		}

		// Update scratchpad
		scratchpad = append(scratchpad,
			fmt.Sprintf("Thought: %s", thought),
			fmt.Sprintf("Action: %s", action),
			fmt.Sprintf("Action Input: %s", actionInput),
			fmt.Sprintf("Observation: %s", observation),
		)
	}

	return nil, fmt.Errorf("max iterations (%d) reached without completion", e.maxIterations)
}

// buildSystemPrompt creates the system prompt with tool descriptions
func (e *ReActEngine) buildSystemPrompt(ctx context.Context) (string, error) {
	resp, err := e.executorClient.ListTools(ctx, &executorpb.ListToolsRequest{})
	if err != nil {
		return "", fmt.Errorf("failed to list tools: %w", err)
	}

	toolDescriptions := []string{}
	for _, tool := range resp.Tools {
		toolDescriptions = append(toolDescriptions,
			fmt.Sprintf("- %s: %s", tool.Name, tool.Description),
		)
	}

	return fmt.Sprintf(`You are an expert Business Analyst AI assistant. Your task is to help analyze requirements and generate documentation.

You have access to the following tools:
%s

Use the following format:

Thought: Analyze the current situation and decide what to do next
Action: tool_name
Action Input: {"key": "value"}
Observation: [Tool result will appear here]
... (repeat Thought/Action/Observation as needed)
Thought: I now have enough information to provide the final answer
Final Answer: [Your comprehensive response]

IMPORTANT:
- Always think step by step
- Use tools when you need information or to perform actions
- Provide detailed, well-structured final answers
- Use Mermaid diagrams when appropriate`, strings.Join(toolDescriptions, "\n")), nil
}

// buildPrompt constructs the full prompt with task and scratchpad
func (e *ReActEngine) buildPrompt(task string, context map[string]string, scratchpad []string, systemPrompt string) string {
	contextStr := ""
	if len(context) > 0 {
		contextJSON, _ := json.Marshal(context)
		contextStr = fmt.Sprintf("\nContext: %s", string(contextJSON))
	}

	scratchpadStr := ""
	if len(scratchpad) > 0 {
		scratchpadStr = "\n\n" + strings.Join(scratchpad, "\n")
	}

	return fmt.Sprintf("%s\n\nTask: %s%s%s\n\nThought:", systemPrompt, task, contextStr, scratchpadStr)
}

// callLLM makes a request to the AI Proxy
func (e *ReActEngine) callLLM(ctx context.Context, prompt string) (*LLMResponse, error) {
	resp, err := e.aiProxyClient.Complete(ctx, &aiproxy.CompleteRequest{
		Payload: &aiproxy.CompletePayload{
			ModelId:     e.modelID,
			Prompt:      prompt,
			Temperature: 0.7,
			MaxTokens:   4096,
		},
	})

	if err != nil {
		return nil, err
	}

	if resp.Result.Code != aiproxy.ResultCode_SUCCESS {
		return nil, fmt.Errorf("AI Proxy error: %s", resp.Result.Message)
	}

	return &LLMResponse{
		Text:       resp.Completion.Text,
		TokensUsed: int(resp.Completion.TotalTokens),
		Cost:       float32(resp.Completion.TotalTokens) * 0.00001, // Rough estimate
	}, nil
}

// LLMResponse contains the LLM response data
type LLMResponse struct {
	Text       string
	TokensUsed int
	Cost       float32
}

// parseResponse extracts thought, action, and action input from LLM response
func (e *ReActEngine) parseResponse(text string) (thought, action, actionInput string, isFinal bool) {
	// Check for Final Answer
	if strings.Contains(text, "Final Answer:") {
		thought = e.extractSection(text, "Thought:", "Final Answer:")
		return thought, "", "", true
	}

	// Extract sections
	thought = e.extractSection(text, "Thought:", "Action:")
	action = e.extractSection(text, "Action:", "Action Input:")
	actionInput = e.extractSection(text, "Action Input:", "Observation:")

	// Clean up
	action = strings.TrimSpace(action)
	actionInput = strings.TrimSpace(actionInput)

	return thought, action, actionInput, false
}

// extractSection extracts text between two markers
func (e *ReActEngine) extractSection(text, start, end string) string {
	startIdx := strings.Index(text, start)
	if startIdx == -1 {
		return ""
	}
	startIdx += len(start)

	endIdx := strings.Index(text[startIdx:], end)
	if endIdx == -1 {
		return strings.TrimSpace(text[startIdx:])
	}

	return strings.TrimSpace(text[startIdx : startIdx+endIdx])
}

// extractFinalAnswer extracts the final answer from the response
func (e *ReActEngine) extractFinalAnswer(text string) string {
	finalAnswerIdx := strings.Index(text, "Final Answer:")
	if finalAnswerIdx == -1 {
		return text
	}

	return strings.TrimSpace(text[finalAnswerIdx+len("Final Answer:"):])
}
