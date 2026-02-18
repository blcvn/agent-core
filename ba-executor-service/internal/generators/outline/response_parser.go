package outline

import (
	"regexp"
	"strings"
	"unicode"

	v32 "github.com/blcvn/backend/services/pkg/domain/v3.2"
)

// ResponseParser parses the markdown response from the LLM into a structured URDOutline
type ResponseParser struct{}

// NewResponseParser creates a new response parser
func NewResponseParser() *ResponseParser {
	return &ResponseParser{}
}

// ParseOutline parses the markdown string into a domain.URDOutline struct
func (p *ResponseParser) ParseOutline(markdown string) (*v32.URDOutline, error) {
	outline := &v32.URDOutline{
		UseCases: []v32.UseCase{},
	}

	// Extract Use Cases
	outline.UseCases = p.parseUseCases(markdown)

	return outline, nil
}

func (p *ResponseParser) parseUseCases(markdown string) []v32.UseCase {
	var useCases []v32.UseCase
	// Split by "## UC-" to find each use case section
	sections := strings.Split(markdown, "## UC-")
	for i, section := range sections {
		if i == 0 {
			continue // Header before first UC
		}

		fullSection := "UC-" + section
		lines := strings.Split(fullSection, "\n")
		if len(lines) == 0 {
			continue
		}
		header := lines[0]

		reID := regexp.MustCompile(`(UC-\d+)`)
		matchID := reID.FindString(header)
		if matchID == "" {
			continue
		}

		name := strings.TrimPrefix(header, matchID+":")
		name = strings.TrimPrefix(name, matchID)
		name = strings.TrimSpace(name)
		if strings.HasPrefix(name, ":") {
			name = strings.TrimSpace(name[1:])
		}

		uc := v32.UseCase{
			ID:   matchID,
			Name: name,
		}

		// Extract actors, conditions, flows using helpers
		uc.Description = p.extractBlock(fullSection, "**Tóm tắt (Brief Description):**")
		uc.PrimaryActor = p.extractActor(fullSection, "- **Sơ cấp (Primary):**")
		uc.Preconditions = p.extractList(fullSection, "**Điều kiện tiên quyết (Preconditions):**")
		uc.Postconditions = p.extractList(fullSection, "**Kết quả mong đợi (Postconditions):**")
		uc.MainFlow = p.extractList(fullSection, "**Luồng sự kiện chính (Main Flow):**")

		useCases = append(useCases, uc)
	}
	return useCases
}

func (p *ResponseParser) extractBlock(section, marker string) string {
	idx := strings.Index(section, marker)
	if idx == -1 {
		return ""
	}
	content := section[idx+len(marker):]
	lines := strings.Split(content, "\n")
	var result []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if (strings.HasPrefix(trimmed, "**") && !strings.Contains(trimmed, marker)) || strings.HasPrefix(trimmed, "---") || strings.HasPrefix(trimmed, "#") {
			break
		}
		result = append(result, trimmed)
	}
	return strings.Join(result, " ")
}

func (p *ResponseParser) extractActor(section, marker string) string {
	idx := strings.Index(section, marker)
	if idx == -1 {
		return ""
	}
	content := section[idx+len(marker):]
	lines := strings.Split(content, "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0])
	}
	return ""
}

func (p *ResponseParser) extractList(section, marker string) []string {
	idx := strings.Index(section, marker)
	if idx == -1 {
		return nil
	}
	content := section[idx+len(marker):]
	lines := strings.Split(content, "\n")
	var result []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		// Stop if we hit next section or specific marker
		if (strings.HasPrefix(trimmed, "**") && !strings.Contains(trimmed, marker)) || strings.HasPrefix(trimmed, "---") || strings.HasPrefix(trimmed, "#") {
			break
		}
		if strings.HasPrefix(trimmed, "-") || strings.HasPrefix(trimmed, "*") {
			result = append(result, strings.TrimSpace(trimmed[1:]))
		} else if len(trimmed) > 0 && unicode.IsDigit(rune(trimmed[0])) && strings.Contains(trimmed, ".") {
			// Numeric list
			dotIdx := strings.Index(trimmed, ".")
			result = append(result, strings.TrimSpace(trimmed[dotIdx+1:]))
		}
	}
	return result
}
