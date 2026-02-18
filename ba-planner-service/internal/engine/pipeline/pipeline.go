package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/blcvn/backend/services/ba-agent-service/domain"
	"github.com/blcvn/backend/services/ba-agent-service/engine/stages/generation"
	"github.com/blcvn/backend/services/ba-agent-service/engine/stages/normalization"
	"github.com/blcvn/backend/services/ba-agent-service/engine/stages/reasoning"
	"github.com/blcvn/backend/services/ba-agent-service/infrastructure/llm"
	"github.com/blcvn/backend/services/ba-agent-service/infrastructure/prompt"
	"github.com/blcvn/backend/services/ba-agent-service/tools"

	proxy_pb "github.com/blcvn/kratos-proto/go/ai-proxy"
	baagent "github.com/blcvn/kratos-proto/go/ba-agent"
	prompt_pb "github.com/blcvn/kratos-proto/go/prompt"
	"github.com/go-redis/redis/v8"
)

// PipelineEngine implements a sequential workflow engine
type PipelineEngine struct {
	aiProxyClient proxy_pb.AIProxyServiceClient
	promptClient  prompt_pb.PromptServiceClient
	toolManager   *tools.ToolManager
	redisClient   *redis.Client
	modelID       string
}

// NewPipelineEngine creates a new Pipeline engine
func NewPipelineEngine(
	aiProxyClient proxy_pb.AIProxyServiceClient,
	promptClient prompt_pb.PromptServiceClient,
	toolManager *tools.ToolManager,
	redisClient *redis.Client,
	modelID string,
) *PipelineEngine {
	return &PipelineEngine{
		aiProxyClient: aiProxyClient,
		promptClient:  promptClient,
		toolManager:   toolManager,
		redisClient:   redisClient,
		modelID:       modelID,
	}
}

// Execute runs the pipeline workflow
func (e *PipelineEngine) Execute(ctx context.Context, task string, payload *baagent.ExecuteTaskPayload) (*ExecuteResult, error) {
	log.Println("Starting Pipeline Execution...")
	startTime := time.Now()

	// 0. Initialize Adapaters
	rawLLM := llm.NewAIProxyAdapter(e.aiProxyClient, e.modelID)
	llmAdapter := llm.NewCachedLLM(rawLLM, e.redisClient, 24*time.Hour, e.modelID)
	promptAdapter := prompt.NewPromptAdapter(e.promptClient, "templates/prompts")

	// 1. Prepare Input
	// Map ExecuteTaskPayload to domain.AgentInput
	// Use TaskDescription as content
	content := task

	// Create Ingestion Payload structure
	// Normalization expects IngestionData JSON with Blocks
	// We wrap the raw text into a single block for now (simplest ingestion)
	ingestionData := domain.IngestionData{
		Blocks: []domain.Block{
			{Type: "text", Text: content, Position: 0},
		},
	}
	ingestionBytes, _ := json.Marshal(ingestionData)

	normInput := domain.AgentInput{
		JobID:     "temp-job-id", // TODO: use actual ID if available
		InputType: "markdown",    // Assumption for raw text
		Source:    "user-input",
		Payload:   ingestionBytes,
	}

	// 2. Normalization Stage
	normAgent := normalization.NewNormalizationAgent(llmAdapter, promptAdapter)
	normOut, err := normAgent.Execute(normInput)
	if err != nil {
		return e.errorResult(err, startTime), nil
	}
	if !normOut.Success {
		return e.errorResult(fmt.Errorf("%s", normOut.Error), startTime), nil
	}

	var canonicalPRD domain.CanonicalPRD
	if err := json.Unmarshal(normOut.Payload, &canonicalPRD); err != nil {
		return e.errorResult(fmt.Errorf("failed to unmarshal normalization result: %w", err), startTime), nil
	}
	log.Printf("Pipeline: Normalization complete. Title: %s", canonicalPRD.Title)

	// 3. Reasoning Stage
	reasoningAgent := reasoning.NewReasoningAgent(llmAdapter, promptAdapter)
	graph, err := reasoningAgent.Execute(canonicalPRD)
	if err != nil {
		return e.errorResult(fmt.Errorf("reasoning failed: %w", err), startTime), nil
	}
	log.Printf("Pipeline: Reasoning complete. Nodes: %d", len(graph.Nodes))

	// 4. Generation Stage
	// For template path, we point to the local file we copied
	// In production, this might be configured via ENV
	tmplPath := "templates/urd_phase4.tmpl"
	genAgent := generation.NewURDGenerationAgent(tmplPath)

	genOut, err := genAgent.Execute(graph)
	if err != nil {
		return e.errorResult(fmt.Errorf("generation failed: %w", err), startTime), nil
	}
	if !genOut.Success {
		return e.errorResult(fmt.Errorf("%s", genOut.Error), startTime), nil
	}

	finalOutput := string(genOut.Payload)
	log.Printf("Pipeline: Generation complete. Length: %d", len(finalOutput))

	return &ExecuteResult{
		FinalAnswer:     finalOutput,
		Steps:           []*ReActStep{}, // Pipeline doesn't produce ReAct steps yet
		TotalIterations: 1,
		ToolCalls:       0,
		LLMCalls:        3, // Approx
		DurationMs:      time.Since(startTime).Milliseconds(),
		TotalCost:       0,
		TotalTokens:     0,
		Status:          "COMPLETED",
	}, nil
}

func (e *PipelineEngine) errorResult(err error, startTime time.Time) *ExecuteResult {
	return &ExecuteResult{
		FinalAnswer: fmt.Sprintf("Pipeline Error: %v", err),
		Status:      "FAILED",
		DurationMs:  time.Since(startTime).Milliseconds(),
	}
}
