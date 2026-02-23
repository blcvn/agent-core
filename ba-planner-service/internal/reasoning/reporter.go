package chat_agent

import (
	"context"
	"fmt"
	"time"

	v32 "github.com/blcvn/backend/services/pkg/domain/v3.2"
)

// ChangeReporter generates reports for document changes
type ChangeReporter struct {
}

// NewChangeReporter creates a new ChangeReporter
func NewChangeReporter() *ChangeReporter {
	return &ChangeReporter{}
}

// GenerateReport generates a change report
func (r *ChangeReporter) GenerateReport(ctx context.Context, oldDoc, newDoc *v32.Document, modifications []Modification, comment string) (*ChangeReport, error) {
	summary := ChangeSummary{
		TotalModifications:   len(modifications),
		AffectedSectionTypes: make(map[string]int),
	}

	var details []ChangeDetail
	var impact ImpactAnalysis

	for _, mod := range modifications {
		// Update summary counts
		switch mod.ActionType {
		case "add":
			summary.AddedSections++
		case "modify":
			summary.ModifiedSections++
		case "delete":
			summary.DeletedSections++
		}
		summary.AffectedSectionTypes[mod.SectionType]++

		// Create detail
		detail := ChangeDetail{
			ModificationID: mod.ModificationID,
			SectionType:    mod.SectionType,
			SectionID:      mod.SectionID,
			ActionType:     mod.ActionType,
			BeforeSnippet:  truncate(mod.OldContent, 200),
			AfterSnippet:   truncate(mod.NewContent, 200),
			Reasoning:      mod.Reasoning,
		}
		details = append(details, detail)

		// Aggregate impact
		// (Simplified logic here)
		if mod.SectionType == "use_case" {
			impact.AffectedUseCases = append(impact.AffectedUseCases, mod.SectionID)
		}
	}

	report := &ChangeReport{
		ReportID:          fmt.Sprintf("REPORT-%d", time.Now().UnixNano()),
		Timestamp:         time.Now(),
		TriggeringComment: comment,
		OldVersion:        oldDoc.Version,
		NewVersion:        newDoc.Version,
		Summary:           summary,
		DetailedChanges:   details,
		ImpactAnalysis:    impact,
		DiffMarkdown:      generateDiffMarkdown(modifications),
	}

	return report, nil
}

func truncate(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen] + "..."
	}
	return s
}

func generateDiffMarkdown(mods []Modification) string {
	diff := "# Change Report\n\n"
	for _, mod := range mods {
		diff += fmt.Sprintf("## %s: %s\n**Action:** %s\n\n", mod.SectionType, mod.SectionID, mod.ActionType)
		if mod.ActionType == "modify" || mod.ActionType == "delete" {
			diff += fmt.Sprintf("**Before:**\n```\n%s\n```\n", mod.OldContent)
		}
		if mod.ActionType == "modify" || mod.ActionType == "add" {
			diff += fmt.Sprintf("**After:**\n```\n%s\n```\n", mod.NewContent)
		}
		diff += fmt.Sprintf("**Reasoning:** %s\n\n---\n\n", mod.Reasoning)
	}
	return diff
}
