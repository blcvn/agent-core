package full

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/blcvn/backend/services/pkg/domain"
	v32 "github.com/blcvn/backend/services/pkg/domain/v3.2"
	"github.com/blcvn/backend/services/pkg/infrastructure/prompt"
	"golang.org/x/sync/errgroup"
)

// GeneratorV2 implements domain.FullGeneratorV2
type GeneratorV2 struct {
	llm           domain.LLMService
	promptAdapter *prompt.PromptAdapter
}

// NewGeneratorV2 creates a new full generator V2
func NewGeneratorV2(llm domain.LLMService, promptAdapter *prompt.PromptAdapter) *GeneratorV2 {
	return &GeneratorV2{
		llm:           llm,
		promptAdapter: promptAdapter,
	}
}

// GenerateFull generates a comprehensive URD from Knowledge Graph context
func (g *GeneratorV2) GenerateFull(ctx context.Context, kg *v32.RequirementGraph, moduleName string, previousContent string) (string, error) {
	log.Printf("[FullGeneratorV2] Starting Full URD generation for module: %s", moduleName)

	// Step 1: Extract context from KG
	extractor := NewContextExtractor()
	moduleCtx, err := extractor.ExtractModuleContext(ctx, kg, moduleName)
	if err != nil {
		log.Printf("[FullGeneratorV2] Failed to extract module context: %v", err)
		return "", fmt.Errorf("failed to extract module context: %w", err)
	}

	// Step 2: Build prompts & Generate concurrently
	// Fetch templates from Prompt Service
	vars := map[string]string{
		"ModuleName":  moduleName,
		"CurrentDate": time.Now().Format("2006-01-02"),
	}

	// System prompt (shared)
	systemPrompt, err := g.promptAdapter.GetPrompt(ctx, "ba_agent_full_system", nil)
	if err != nil {
		return "", fmt.Errorf("failed to get system prompt: %w", err)
	}

	// Instruction prompt (template for output structure)
	instructionPrompt, err := g.promptAdapter.GetPrompt(ctx, "ba_agent_full_instruction", vars)
	if err != nil {
		return "", fmt.Errorf("failed to get instruction prompt: %w", err)
	}

	promptBuilder := NewPromptBuilder()

	// Initialize concurrent generation
	gGroup, gCtx := errgroup.WithContext(ctx)

	var part1Overview, part2UseCases, part3Technical string

	// PART 1: Overview
	gGroup.Go(func() error {
		log.Printf("[FullGeneratorV2] Generating Part 1: Overview...")
		p1UserPrompt := promptBuilder.BuildOverviewPrompt(moduleCtx, instructionPrompt, previousContent)
		resp, err := g.llm.Chat(gCtx, systemPrompt, p1UserPrompt)
		if err != nil {
			return fmt.Errorf("failed to generate overview: %w", err)
		}
		part1Overview = g.cleanMarkdown(resp)
		log.Printf("[FullGeneratorV2] Part 1: Overview generated (%d chars)", len(part1Overview))
		return nil
	})

	// PART 2: Use Cases
	gGroup.Go(func() error {
		log.Printf("[FullGeneratorV2] Generating Part 2: Use Cases...")
		p2UserPrompt := promptBuilder.BuildUseCasesPrompt(moduleCtx, instructionPrompt, previousContent)
		resp, err := g.llm.Chat(gCtx, systemPrompt, p2UserPrompt)
		if err != nil {
			return fmt.Errorf("failed to generate use cases: %w", err)
		}
		part2UseCases = g.cleanMarkdown(resp)
		log.Printf("[FullGeneratorV2] Part 2: Use Cases generated (%d chars)", len(part2UseCases))
		return nil
	})

	// PART 3: Technical
	gGroup.Go(func() error {
		log.Printf("[FullGeneratorV2] Generating Part 3: Technical...")
		p3UserPrompt := promptBuilder.BuildTechnicalPrompt(moduleCtx, instructionPrompt, previousContent)
		resp, err := g.llm.Chat(gCtx, systemPrompt, p3UserPrompt)
		if err != nil {
			return fmt.Errorf("failed to generate technical section: %w", err)
		}
		part3Technical = g.cleanMarkdown(resp)
		log.Printf("[FullGeneratorV2] Part 3: Technical generated (%d chars)", len(part3Technical))
		return nil
	})

	// Wait for all parts
	if err := gGroup.Wait(); err != nil {
		log.Printf("[FullGeneratorV2] Concurrent generation failed: %v", err)
		return "", err
	}

	// Step 3: Assembly
	log.Printf("[FullGeneratorV2] Assembling final URD...")
	finalURD := fmt.Sprintf("# URD - %s\n\n> **Module:** %s\n> **Version:** 1.0\n> **Created Date:** %s\n\n%s\n\n%s\n\n%s",
		moduleName, moduleName, time.Now().Format("2006-01-02"),
		part1Overview, part2UseCases, part3Technical)

	log.Printf("[FullGeneratorV2] Full URD generation completed successfully. Total length: %d", len(finalURD))
	return finalURD, nil
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
