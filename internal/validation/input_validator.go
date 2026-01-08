package validation

import (
	"fmt"
	"regexp"
	"strings"
)

// InputValidator detects prompt injection attacks and malicious requests.
type InputValidator struct {
	promptInjectionPatterns []*regexp.Regexp
	jailbreakPatterns       []*regexp.Regexp
}

// NewInputValidator creates a new validator with precompiled patterns.
func NewInputValidator() *InputValidator {
	return &InputValidator{
		promptInjectionPatterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)(ignore|forget|override|new\s+role|you\s+are\s+now).*previous`),
			regexp.MustCompile(`(?i)(system|user)\s*prompt`),
			regexp.MustCompile(`(?i)(repeat|show|display|print).*instruction`),
			regexp.MustCompile(`(?i)(here\s+are?\s+your?\s+)?instructions?`),
			regexp.MustCompile(`(?i)(from\s+now|starting\s+now).*do\s+not`),
			regexp.MustCompile(`(?i)roleplay.*(developer|admin|root)`),
		},
		jailbreakPatterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)(dan|developer|jailbreak|uncensored)`),
			regexp.MustCompile(`(?i)(free|unrestricted|no\s+rules)`),
			regexp.MustCompile(`(?i)(hacker|exploit|vulnerability)`),
			regexp.MustCompile(`(?i)(sql|select.*from|drop\s+table)`),
			regexp.MustCompile(`(?i)(javascript|script|eval|exec)`),
		},
	}
}

// ValidateInput checks for prompt injection and jailbreak attempts.
func (v *InputValidator) ValidateInput(input string) error {
	lowerInput := strings.ToLower(input)

	// Check for prompt injection patterns
	for _, pattern := range v.promptInjectionPatterns {
		if pattern.MatchString(lowerInput) {
			return fmt.Errorf("prompt injection detected")
		}
	}

	// Tier 2: Jailbreak and exploit patterns
	for _, pattern := range v.jailbreakPatterns {
		if pattern.MatchString(lowerInput) {
			return fmt.Errorf("jailbreak attempt detected")
		}
	}

	// Tier 3: Length and content heuristics
	if len(input) > 5000 {
		return fmt.Errorf("input too long")
	}

	if strings.Contains(lowerInput, "```") && strings.Contains(lowerInput, "go ") {
		// Potential code injection attempt
		return fmt.Errorf("code injection attempt detected")
	}

	return nil
}
