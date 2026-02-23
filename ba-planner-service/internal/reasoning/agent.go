package chat_agent

import (
	"context"
	"fmt"

	"github.com/blcvn/backend/services/pkg/domain"
	v32 "github.com/blcvn/backend/services/pkg/domain/v3.2"
)

// ChatAgent orchestrates the interactive document update process
type ChatAgent struct {
	classifier     *IntentClassifier
	extractor      *SectionExtractor
	contextBuilder *ContextBuilder
	generator      *ModificationGenerator
	applier        *ModificationApplier
	kgUpdater      *KGUpdater
	reporter       *ChangeReporter
}

// NewChatAgent creates a new ChatAgent
func NewChatAgent(llmService domain.LLMService) *ChatAgent {
	return &ChatAgent{
		classifier:     NewIntentClassifier(llmService),
		extractor:      NewSectionExtractor(llmService),
		contextBuilder: NewContextBuilder(),
		generator:      NewModificationGenerator(llmService),
		applier:        NewModificationApplier(),
		kgUpdater:      NewKGUpdater(),
		reporter:       NewChangeReporter(),
	}
}

// ProcessCommentAndUpdateDocument executes the full workflow
func (a *ChatAgent) ProcessCommentAndUpdateDocument(ctx context.Context, comment string, doc *v32.Document, kg *v32.RequirementGraph) (*v32.Document, *ChangeReport, error) {
	fmt.Printf("[ChatAgent] Step 1: Classifying Intent...\n")
	intent, err := a.classifier.ClassifyCommentIntent(ctx, comment)
	if err != nil {
		return nil, nil, fmt.Errorf("step 1 failed: %w", err)
	}

	fmt.Printf("[ChatAgent] Step 2: Extracting Affected Sections...\n")
	affectedSections, err := a.extractor.ExtractAffectedSections(ctx, comment, doc, kg)
	if err != nil {
		return nil, nil, fmt.Errorf("step 2 failed: %w", err)
	}

	fmt.Printf("[ChatAgent] Step 3: Building Modification Context...\n")
	modificationContext, err := a.contextBuilder.BuildModificationContext(ctx, affectedSections, doc, kg)
	if err != nil {
		return nil, nil, fmt.Errorf("step 3 failed: %w", err)
	}

	fmt.Printf("[ChatAgent] Step 4: Generating Modifications...\n")
	modifications, err := a.generator.GenerateModifications(ctx, comment, intent, modificationContext, doc.Tier)
	if err != nil {
		return nil, nil, fmt.Errorf("step 4 failed: %w", err)
	}

	fmt.Printf("[ChatAgent] Step 5: Applying Modifications...\n")
	updatedDoc, err := a.applier.ApplyModifications(ctx, doc, modifications)
	if err != nil {
		return nil, nil, fmt.Errorf("step 5 failed: %w", err)
	}

	fmt.Printf("[ChatAgent] Step 6: Updating Knowledge Graph...\n")
	_, err = a.kgUpdater.UpdateKG(ctx, kg, modifications, updatedDoc)
	if err != nil {
		fmt.Printf("[ChatAgent] Warning: KG update failed: %v\n", err)
		// Non-critical (?)
	}

	fmt.Printf("[ChatAgent] Step 7: Generating Change Report...\n")
	report, err := a.reporter.GenerateReport(ctx, doc, updatedDoc, modifications, comment)
	if err != nil {
		return nil, nil, fmt.Errorf("step 7 failed: %w", err)
	}

	// Enrich report with Intent and IdentifiedSections
	report.Intent = intent
	report.IdentifiedSections = affectedSections

	return updatedDoc, report, nil
}
