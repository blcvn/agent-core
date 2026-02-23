package outline

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/blcvn/ba-shared-libs/pkg/domain"
	v32 "github.com/blcvn/ba-shared-libs/pkg/domain/v3.2"
	"github.com/blcvn/ba-shared-libs/pkg/infrastructure/prompt"
)

// GeneratorV2 implements domain.OutlineGeneratorV2
type GeneratorV2 struct {
	llm           domain.LLMService
	promptAdapter *prompt.PromptAdapter
}

// NewGeneratorV2 creates a new outline generator V2
func NewGeneratorV2(llm domain.LLMService, promptAdapter *prompt.PromptAdapter) *GeneratorV2 {
	return &GeneratorV2{
		llm:           llm,
		promptAdapter: promptAdapter,
	}
}

// GenerateOutline generates a URD Outline from Knowledge Graph context
func (g *GeneratorV2) GenerateOutline(ctx context.Context, kg *v32.RequirementGraph, moduleName string, previousContent string) (string, error) {
	log.Printf("[OutlineGeneratorV2] Starting Outline generation for module: %s", moduleName)

	// Step 1: Extract context from KG
	extractor := NewContextExtractor()
	moduleCtx, err := extractor.ExtractModuleContext(ctx, kg, moduleName)
	if err != nil {
		log.Printf("[OutlineGeneratorV2] Failed to extract module context: %v", err)
		return "", fmt.Errorf("failed to extract module context: %w", err)
	}

	// Step 2: Build prompts
	// Fetch templates from Prompt Service
	vars := map[string]string{
		"ModuleName":  moduleName,
		"CurrentDate": time.Now().Format("2006-01-02"),
	}

	systemPrompt, err := g.promptAdapter.GetPrompt(ctx, "ba_agent_outline_system", nil)
	if err != nil {
		return "", fmt.Errorf("failed to get system prompt: %w", err)
	}

	instructionPrompt, err := g.promptAdapter.GetPrompt(ctx, "ba_agent_outline_instruction", vars)
	if err != nil {
		return "", fmt.Errorf("failed to get instruction prompt: %w", err)
	}

	promptBuilder := NewPromptBuilder()
	// Build user prompt with extracted instruction and previous content (Index)
	userPrompt := promptBuilder.BuildUserPrompt(moduleCtx, instructionPrompt, previousContent)

	// Step 3: Call LLM
	log.Printf("[OutlineGeneratorV2] System prompt length: %d", len(systemPrompt))
	log.Printf("[OutlineGeneratorV2] User prompt length: %d", len(userPrompt))
	response, err := g.llm.Chat(ctx, systemPrompt, userPrompt)
	if err != nil {
		log.Printf("[OutlineGeneratorV2] Failed to generate content from LLM: %v", err)
		return "", fmt.Errorf("failed to generate content from LLM: %w", err)
	}

	// Step 4: Clean response
	cleanedResponse := g.cleanMarkdown(response)
	log.Printf("[OutlineGeneratorV2] Outline generation completed successfully.")
	return cleanedResponse, nil
}

func (g *GeneratorV2) cleanMarkdown(content string) string {
	trimmed := strings.TrimSpace(content)
	// Check if it starts with ```
	if strings.HasPrefix(trimmed, "```") {
		// Find the newline after the first ``` line
		// This handles ```markdown, ```go, or just ```
		firstLineEnd := strings.Index(trimmed, "\n")
		if firstLineEnd == -1 {
			// content is just ```...``` (one line? unlikely for md) or ```
			// If it's just the fence, return empty?
			return ""
		}

		// Check if it ends with ```
		if strings.HasSuffix(trimmed, "```") {
			// Find the last instance of ```
			lastTripleQuote := strings.LastIndex(trimmed, "```")
			if lastTripleQuote > firstLineEnd {
				// Return everything between first line and last ```
				return strings.TrimSpace(trimmed[firstLineEnd+1 : lastTripleQuote])
			}
		}
	}
	// If it doesn't match the strict wrapping pattern, return as is.
	// This preserves inner code blocks if the outer content is raw markdown.
	return content
}
