package chat_agent

import (
	"time"

	v32 "github.com/blcvn/ba-shared-libs/pkg/domain/v3.2"
)

// CommentIntent represents the classified intent of a user's comment
type CommentIntent struct {
	Type     string   `json:"type"`     // add, modify, delete, question, clarification
	Scope    string   `json:"scope"`    // specific_section, entire_document, cross_cutting
	Urgency  string   `json:"urgency"`  // low, medium, high
	Keywords []string `json:"keywords"` // extracted keywords
}

// AffectedSection represents a section of the document affected by the comment
type AffectedSection struct {
	SectionType     string   `json:"section_type"` // use_case, actor, entity, integration, business_rule, flow_step, general
	SectionID       string   `json:"section_id"`
	SectionName     string   `json:"section_name"`
	CurrentContent  string   `json:"current_content"`
	ConfidenceScore float64  `json:"confidence_score"`
	Reasoning       string   `json:"reasoning"`
	RelatedNodeIDs  []string `json:"related_node_ids"` // IDs of related KG nodes
}

// ModificationContext represents the context needed to generate modifications
type ModificationContext struct {
	AffectedSections     []AffectedSection  `json:"affected_sections"`
	RelatedUserStories   []v32.UserStory    `json:"related_user_stories"` // Using v32 domain types if available, otherwise define lightweight versions
	RelatedFeatures      []v32.Feature      `json:"related_features"`
	RelatedBusinessRules []v32.BusinessRule `json:"related_business_rules"`
	RelatedIntegrations  []v32.Integration  `json:"related_integrations"`
	DependentSections    []SectionRef       `json:"dependent_sections"`
	ConstraintsFromPRD   []string           `json:"constraints_from_prd"`
	SimilarSections      []SectionRef       `json:"similar_sections"`
}

// SectionRef is a lightweight reference to a section
type SectionRef struct {
	Type    string `json:"type"`
	ID      string `json:"id"`
	Name    string `json:"name"`
	Content string `json:"content,omitempty"`
}

// Modification represents a proposed change to the document
type Modification struct {
	ModificationID string `json:"modification_id"`
	SectionType    string `json:"section_type"`
	SectionID      string `json:"section_id"`
	ActionType     string `json:"action_type"` // add, modify, delete, no_change
	OldContent     string `json:"old_content,omitempty"`
	NewContent     string `json:"new_content,omitempty"`
	Reasoning      string `json:"reasoning"`
	ImpactAnalysis string `json:"impact_analysis"`
}

// ChangeReport represents the summary of changes
type ChangeReport struct {
	ReportID           string            `json:"report_id"`
	Timestamp          time.Time         `json:"timestamp"`
	TriggeringComment  string            `json:"triggering_comment"`
	OldVersion         int               `json:"old_version"`
	NewVersion         int               `json:"new_version"`
	Summary            ChangeSummary     `json:"summary"`
	DetailedChanges    []ChangeDetail    `json:"detailed_changes"`
	ImpactAnalysis     ImpactAnalysis    `json:"impact_analysis"`
	DiffMarkdown       string            `json:"diff_markdown"`
	DiffHTML           string            `json:"diff_html"` // Optional
	Intent             *CommentIntent    `json:"intent"`
	IdentifiedSections []AffectedSection `json:"identified_sections"`
}

type ChangeSummary struct {
	TotalModifications   int            `json:"total_modifications"`
	AddedSections        int            `json:"added_sections"`
	ModifiedSections     int            `json:"modified_sections"`
	DeletedSections      int            `json:"deleted_sections"`
	AffectedSectionTypes map[string]int `json:"affected_section_types"`
}

type ChangeDetail struct {
	ModificationID string `json:"modification_id"`
	SectionType    string `json:"section_type"`
	SectionID      string `json:"section_id"`
	SectionName    string `json:"section_name"`
	ActionType     string `json:"action_type"`
	BeforeSnippet  string `json:"before_snippet"`
	AfterSnippet   string `json:"after_snippet"`
	Reasoning      string `json:"reasoning"`
}

type ImpactAnalysis struct {
	AffectedUseCases  []string `json:"affected_use_cases"`
	AffectedActors    []string `json:"affected_actors"`
	AffectedEntities  []string `json:"affected_entities"`
	DependentSections []string `json:"dependent_sections"`
	ConsistencyIssues []string `json:"consistency_issues"`
}
