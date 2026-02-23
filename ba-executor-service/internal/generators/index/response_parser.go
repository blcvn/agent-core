package index

import (
	"regexp"
	"strings"

	v32 "github.com/blcvn/ba-shared-libs/pkg/domain/v3.2"
)

// ResponseParser parses the markdown response from the LLM into a structured URDIndex
type ResponseParser struct{}

// NewResponseParser creates a new response parser
func NewResponseParser() *ResponseParser {
	return &ResponseParser{}
}

// ParseIndex parses the markdown string into a domain.URDIndex struct
func (p *ResponseParser) ParseIndex(markdown string) (*v32.URDIndex, error) {
	index := &v32.URDIndex{}

	// Extract US to UC Mapping
	index.USToUCMapping = p.parseUSToUCMapping(markdown)

	// Extract Actors
	index.HumanActors = p.parseHumanActors(markdown)
	index.SystemActors = p.parseSystemActors(markdown)

	// Extract Use Case Summary
	index.UseCaseSummaryTable = p.parseUseCaseSummary(markdown)

	// Extract Integration Touchpoints
	index.IntegrationTouchpoints = p.parseIntegrations(markdown)

	// Extract Data Entities
	index.DataEntities = p.parseDataEntities(markdown)

	return index, nil
}

func (p *ResponseParser) parseUSToUCMapping(markdown string) []v32.USToUCMap {
	var mapping []v32.USToUCMap
	// Regex for table rows: | US-XXX | [Story] | UC-XXX | [Rationale] |
	re := regexp.MustCompile(`\|?\s*(US-\d+)\s*\|[^\|]*\|\s*([^\|]+)\s*\|?\s*([^\|]*)?`)
	matches := re.FindAllStringSubmatch(markdown, -1)

	for _, match := range matches {
		if len(match) >= 3 {
			usID := strings.TrimSpace(match[1])
			ucIDsStr := strings.TrimSpace(match[2])
			note := ""
			if len(match) >= 4 {
				note = strings.TrimSpace(match[3])
			}

			// Split UC IDs (could be multiple, comma separated)
			rawUCs := strings.Split(ucIDsStr, ",")
			var ucIDs []string
			for _, r := range rawUCs {
				id := strings.TrimSpace(r)
				if id != "" && strings.HasPrefix(id, "UC-") {
					ucIDs = append(ucIDs, id)
				}
			}

			if len(ucIDs) > 0 {
				mapping = append(mapping, v32.USToUCMap{
					UserStoryID: usID,
					UseCaseIDs:  ucIDs,
					MappingNote: note,
				})
			}
		}
	}
	return mapping
}

func (p *ResponseParser) parseHumanActors(markdown string) []v32.Actor {
	var actors []v32.Actor
	// Extract Section 2.1
	section := p.extractSection(markdown, "2.1 Human Actors")
	if section == "" {
		return actors
	}

	// Regex for Actor table: | ACT-XXX | Name | Role | Responsibilities | UC-XXX |
	re := regexp.MustCompile(`\|?\s*(ACT-\d+)\s*\|\s*([^\|]+)\s*\|\s*([^\|]+)\s*\|\s*([^\|]+)\s*\|?\s*([^\|]*)?`)
	matches := re.FindAllStringSubmatch(section, -1)

	for _, match := range matches {
		actors = append(actors, v32.Actor{
			ID:          strings.TrimSpace(match[1]),
			Name:        strings.TrimSpace(match[2]),
			Role:        strings.TrimSpace(match[3]),
			Type:        "Human",                     // Hardcoded or use constant if available. Reference used domain.ActorTypeHuman
			Description: strings.TrimSpace(match[4]), // Mapping responsibilities to description for now
		})
	}
	return actors
}

func (p *ResponseParser) parseSystemActors(markdown string) []v32.Actor {
	var actors []v32.Actor
	// Extract Section 2.2
	section := p.extractSection(markdown, "2.2 System Actors")
	if section == "" {
		return actors
	}

	re := regexp.MustCompile(`\|?\s*(ACT-\d+)\s*\|\s*([^\|]+)\s*\|\s*([^\|]+)\s*\|\s*([^\|]+)\s*\|?\s*([^\|]*)?`)
	matches := re.FindAllStringSubmatch(section, -1)

	for _, match := range matches {
		actors = append(actors, v32.Actor{
			ID:          strings.TrimSpace(match[1]),
			Name:        strings.TrimSpace(match[2]),
			Type:        "System", // Hardcoded or use constant
			Role:        strings.TrimSpace(match[3]),
			Description: strings.TrimSpace(match[4]),
		})
	}
	return actors
}

func (p *ResponseParser) parseUseCaseSummary(markdown string) []v32.UseCaseSummary {
	var summaries []v32.UseCaseSummary
	section := p.extractSection(markdown, "4 Use Case Summary Table")
	if section == "" {
		return summaries
	}

	// | UC ID | Name | Actor | Trigger | Outcome/Pre/Post | Priority |
	re := regexp.MustCompile(`\|?\s*(UC-\d+)\s*\|\s*([^\|]+)\s*\|\s*([^\|]+)\s*\|\s*([^\|]+)\s*\|\s*([^\|]+)\s*\|\s*([^\|]+)\s*\|?`)
	matches := re.FindAllStringSubmatch(section, -1)

	for _, match := range matches {
		summaries = append(summaries, v32.UseCaseSummary{
			ID:           strings.TrimSpace(match[1]),
			Name:         strings.TrimSpace(match[2]),
			PrimaryActor: strings.TrimSpace(match[3]),
			Trigger:      strings.TrimSpace(match[4]),
			Priority:     strings.TrimSpace(match[6]),
		})
	}
	return summaries
}

func (p *ResponseParser) parseIntegrations(markdown string) []v32.IntegrationTouchpoint {
	var integrations []v32.IntegrationTouchpoint
	section := p.extractSection(markdown, "6 Integration Touchpoints")
	if section == "" {
		return integrations
	}

	// | INT-XXX | System | Type | Direction | Purpose | UCs |
	re := regexp.MustCompile(`\|?\s*(INT-\d+)\s*\|\s*([^\|]+)\s*\|\s*([^\|]+)\s*\|\s*([^\|]+)\s*\|\s*([^\|]+)\s*\|?`)
	matches := re.FindAllStringSubmatch(section, -1)

	for _, match := range matches {
		integrations = append(integrations, v32.IntegrationTouchpoint{
			ID:          strings.TrimSpace(match[1]),
			Name:        strings.TrimSpace(match[2]),
			Type:        strings.TrimSpace(match[3]),
			Direction:   strings.TrimSpace(match[4]),
			Description: strings.TrimSpace(match[5]),
		})
	}
	return integrations
}

func (p *ResponseParser) parseDataEntities(markdown string) []v32.DataEntity {
	var entities []v32.DataEntity
	section := p.extractSection(markdown, "7 Data Entity Overview")
	if section == "" {
		return entities
	}

	// | ENT-XXX | Name | Description | Attributes | UCs |
	re := regexp.MustCompile(`\|?\s*(ENT-\d+)\s*\|\s*([^\|]+)\s*\|\s*([^\|]+)\s*\|\s*([^\|]+)\s*\|?`)
	matches := re.FindAllStringSubmatch(section, -1)

	for _, match := range matches {
		attrStr := strings.TrimSpace(match[4])
		attrs := strings.Split(attrStr, ",")
		for i := range attrs {
			attrs[i] = strings.TrimSpace(attrs[i])
		}

		entities = append(entities, v32.DataEntity{
			ID:            strings.TrimSpace(match[1]),
			Name:          strings.TrimSpace(match[2]),
			Description:   strings.TrimSpace(match[3]),
			KeyAttributes: attrs,
		})
	}
	return entities
}

func (p *ResponseParser) extractSection(markdown, title string) string {
	lines := strings.Split(markdown, "\n")
	var sectionLines []string
	found := false

	for _, line := range lines {
		if strings.Contains(line, "#") && strings.Contains(line, title) {
			found = true
			continue
		}
		if found && strings.HasPrefix(line, "#") {
			break
		}
		if found {
			sectionLines = append(sectionLines, line)
		}
	}

	return strings.Join(sectionLines, "\n")
}
