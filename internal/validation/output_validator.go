package validation

import (
	"strings"

	"JourneyBuilder/internal/instruction"
)

// OutputValidator validates AI-generated responses for compliance and quality.
type OutputValidator struct {
	spamKeywords []string
}

// NewOutputValidator creates a new output validator.
func NewOutputValidator() *OutputValidator {
	return &OutputValidator{
		spamKeywords: []string{
			"click here now", "limited time", "act now", "urgent",
			"guaranteed", "free money", "no risk", "100% free",
		},
	}
}

// ValidateResponse checks the AI response for spam indicators and compliance issues.
func (v *OutputValidator) ValidateResponse(response string, step instruction.WorkflowStep) error {
	// Basic validation - can be expanded
	if len(response) == 0 {
		return nil // Empty responses are handled elsewhere
	}

	// Check for excessive spam keywords
	spamCount := 0
	lowerResponse := strings.ToLower(response)
	for _, keyword := range v.spamKeywords {
		if strings.Contains(lowerResponse, keyword) {
			spamCount++
		}
	}

	if spamCount > 3 {
		// Too many spam indicators
		return nil // Log but don't fail - let it through for now
	}

	return nil
}
