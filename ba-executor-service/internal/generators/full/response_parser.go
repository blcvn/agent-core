package full

import (
	"regexp"
	"strings"

	v32 "github.com/blcvn/backend/services/pkg/domain/v3.2"
)

// ResponseParser parses Full URD markdown response
type ResponseParser struct {
	ucRegex     *regexp.Regexp
	entityRegex *regexp.Regexp
	apiRegex    *regexp.Regexp
}

// NewResponseParser creates a new response parser
func NewResponseParser() *ResponseParser {
	return &ResponseParser{
		// Match ## 2.1 UC-001: Login
		ucRegex: regexp.MustCompile(`(?m)^##\s+\d+\.\d+\s+(UC-[A-Z0-9-]+):\s*(.+)$`),
		// Match ## 3.1 ENT-001: User
		entityRegex: regexp.MustCompile(`(?m)^##\s+\d+\.\d+\s+(ENT-[A-Z0-9-]+):\s*(.+)$`),
		// Match ## 4.1 INT-001: Payment Gateway
		apiRegex: regexp.MustCompile(`(?m)^##\s+\d+\.\d+\s+(INT-[A-Z0-9-]+):\s*(.+)$`),
	}
}

// ParseFullURD parses the markdown content into a FullURD structure
func (p *ResponseParser) ParseFullURD(content string) (*v32.URDFull, error) {
	fullURD := &v32.URDFull{
		UseCases:          []v32.FullUseCase{},
		APISpecifications: []v32.APISpecification{},
		Entities:          []v32.OutlineEntity{},
	}

	lines := strings.Split(content, "\n")
	var currentSection string
	var currentObject interface{}

	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		// Detect Use Case
		if matches := p.ucRegex.FindStringSubmatch(line); len(matches) > 2 {
			currentSection = "USE_CASE"
			uc := v32.FullUseCase{
				ID:       matches[1],
				Name:     matches[2],
				MainFlow: []v32.DetailedFlowStep{},
			}
			fullURD.UseCases = append(fullURD.UseCases, uc)
			currentObject = &fullURD.UseCases[len(fullURD.UseCases)-1]
			continue
		}

		// Detect Entity
		if matches := p.entityRegex.FindStringSubmatch(line); len(matches) > 2 {
			currentSection = "ENTITY"
			entity := v32.OutlineEntity{
				EntityID: matches[1],
				Name:     matches[2],
			}
			fullURD.Entities = append(fullURD.Entities, entity)
			currentObject = &fullURD.Entities[len(fullURD.Entities)-1]
			continue
		}

		// Detect Integration/API
		if matches := p.apiRegex.FindStringSubmatch(line); len(matches) > 2 {
			currentSection = "API"
			// Create a generic API Spec for this integration
			apiSpec := v32.APISpecification{
				ID:        matches[1],
				Name:      matches[2],
				Endpoints: []v32.APIEndpoint{},
			}
			fullURD.APISpecifications = append(fullURD.APISpecifications, apiSpec)
			currentObject = &fullURD.APISpecifications[len(fullURD.APISpecifications)-1]
			continue
		}

		// Parse content based on current section
		if currentSection == "USE_CASE" && strings.HasPrefix(line, "###") {
			// Subsections like Description, Flow...
			if strings.Contains(line, "Mô tả") || strings.Contains(line, "Description") {
				// Read next lines until next header
				desc := []string{}
				for j := i + 1; j < len(lines); j++ {
					if strings.HasPrefix(lines[j], "#") {
						break
					}
					if strings.TrimSpace(lines[j]) != "" {
						desc = append(desc, strings.TrimSpace(lines[j]))
					}
					i = j // Advance outer loop
				}
				if uc, ok := currentObject.(*v32.FullUseCase); ok {
					uc.Description = strings.Join(desc, "\n")
				}
			} else if strings.Contains(line, "Luồng sự kiện") || strings.Contains(line, "Flow of Events") {
				// Enhanced parsing for MainFlow
				for j := i + 1; j < len(lines); j++ {
					l := strings.TrimSpace(lines[j])
					if strings.HasPrefix(l, "#") {
						break
					}
					if l != "" {
						// Rudimentary step parsing: check for number
						// regex scan for "1. Step..."
						// For now, just create a step with the whole line as Action
						step := v32.DetailedFlowStep{
							StepNumber: len(currentObject.(*v32.FullUseCase).MainFlow) + 1,
							Action:     l,
						}
						// If line starts with "1. ", strip it
						reStep := regexp.MustCompile(`^\d+\.\s+`)
						step.Action = reStep.ReplaceAllString(step.Action, "")

						if uc, ok := currentObject.(*v32.FullUseCase); ok {
							uc.MainFlow = append(uc.MainFlow, step)
						}
					}
					i = j
				}
			}
		} else if currentSection == "API" {
			// Check for method/path pattern: - **GET /path**
			if strings.HasPrefix(line, "- **") {
				// Regex to capture method and path: - \*\*(GET|POST|PUT|DELETE|PATCH)\s+(.+?)\*\*
				apiLineRegex := regexp.MustCompile(`(?i)\*\*(GET|POST|PUT|DELETE|PATCH)\s+(.+?)\*\*`)
				if matches := apiLineRegex.FindStringSubmatch(line); len(matches) > 2 {
					if apiSpec, ok := currentObject.(*v32.APISpecification); ok {
						endpoint := v32.APIEndpoint{
							Method: strings.ToUpper(matches[1]),
							Path:   matches[2],
						}
						// Check if description follows on same line
						parts := strings.Split(line, ":")
						if len(parts) > 1 {
							endpoint.Description = strings.TrimSpace(parts[1])
						}
						apiSpec.Endpoints = append(apiSpec.Endpoints, endpoint)
					}
				}
			}
		}
	}

	return fullURD, nil
}
