package index

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/blcvn/backend/services/pkg/domain"
	v32 "github.com/blcvn/backend/services/pkg/domain/v3.2"
	"github.com/blcvn/backend/services/pkg/infrastructure/prompt"
)

// GeneratorV2 is the main URD Index generator using KG context
type GeneratorV2 struct {
	llm           domain.LLMService
	promptAdapter *prompt.PromptAdapter
}

// NewGeneratorV2 creates a new index generator V2
func NewGeneratorV2(llm domain.LLMService, promptAdapter *prompt.PromptAdapter) *GeneratorV2 {
	return &GeneratorV2{
		llm:           llm,
		promptAdapter: promptAdapter,
	}
}

// GenerateIndex generates a URD Index from Knowledge Graph context
func (g *GeneratorV2) GenerateIndex(ctx context.Context, kg *v32.RequirementGraph, moduleName string) (string, error) {
	log.Printf("[GeneratorV2] Starting Index generation for module: %s", moduleName)

	// Step 1: Extract context from KG
	log.Printf("[GeneratorV2] Extracting module context from KG...")
	extractor := NewContextExtractor()
	moduleContext, err := extractor.ExtractModuleContext(ctx, kg, moduleName)
	if err != nil {
		log.Printf("[GeneratorV2] Failed to extract module context: %v", err)
		return "", fmt.Errorf("failed to extract module context: %w", err)
	}
	log.Printf("[GeneratorV2] Context extracted: %d features, %d user stories, %d personas",
		len(moduleContext.Features), len(moduleContext.UserStories), len(moduleContext.Personas))

	// Step 2: Build prompts
	log.Printf("[GeneratorV2] Building prompts...")

	// Fetch templates from Prompt Service
	// Variables for instruction template
	vars := map[string]string{
		"ModuleName":  moduleName,
		"CurrentDate": time.Now().Format("2006-01-02"),
	}

	systemPrompt, err := g.promptAdapter.GetPrompt(ctx, "ba_agent_index_system", nil)
	if err != nil {
		return "", fmt.Errorf("failed to get system prompt: %w", err)
	}

	instructionPrompt, err := g.promptAdapter.GetPrompt(ctx, "ba_agent_index_instruction", vars)
	if err != nil {
		return "", fmt.Errorf("failed to get instruction prompt: %w", err)
	}

	promptBuilder := NewPromptBuilder()
	// System prompt is now fetched

	// Build user prompt with extracted instruction
	userPrompt := promptBuilder.BuildUserPrompt(moduleContext, instructionPrompt)

	log.Printf("[GeneratorV2] System prompt length: %d", len(systemPrompt))
	log.Printf("[GeneratorV2] User prompt length: %d", len(userPrompt))

	// Step 3: Call LLM
	log.Printf("[GeneratorV2] Calling LLM for Index generation...")
	response, err := g.llm.Chat(ctx, systemPrompt, userPrompt)
	if err != nil {
		log.Printf("[GeneratorV2] Failed to generate content from LLM: %v", err)
		return "", fmt.Errorf("failed to generate content from LLM: %w", err)
	}

	log.Printf("[GeneratorV2] LLM response received, length: %d", len(response))

	// Step 4: Clean and validate response
	indexContent := cleanMarkdown(response)

	if len(indexContent) < 100 {
		return "", fmt.Errorf("generated index is too short (%d chars), likely invalid", len(indexContent))
	}

	log.Printf("[GeneratorV2] Index generation completed successfully %s and clean %s", response, indexContent)
	return indexContent, nil
}

// cleanMarkdown removes code fences and cleans markdown content
func cleanMarkdown(content string) string {
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
