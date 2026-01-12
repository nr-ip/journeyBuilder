package orchestrator

import (
	"regexp"
	"strings"

	"JourneyBuilder/internal/instruction"
	"JourneyBuilder/internal/models"
)

// ContextBuilder is responsible for stateless context extraction from the request payload.
type ContextBuilder struct {
	maxHistoryLength int
}

func NewContextBuilder(maxHistory int) *ContextBuilder {
	return &ContextBuilder{
		maxHistoryLength: maxHistory,
	}
}

// BuildContext constructs the complete user context from the client payload.
// This is STATELESS – all data comes from the request, not from server storage.
func (cb *ContextBuilder) BuildContext(req *models.ChatRequest) *instruction.UserContext {
	ctx := &instruction.UserContext{
		ConversationHistory: req.ConversationHistory,
	}

	// Extract USP from conversation history or current message
	ctx.ExtractedUSP = cb.extractUSP(req)

	// Extract ICP from conversation history or current message
	ctx.ExtractedICP = cb.extractICP(req)

	// Identify vertical based on context
	ctx.IdentifiedVertical = cb.identifyVertical(req)

	// Extract Circle of Trust if mentioned
	ctx.CurrentCircleOfTrust = cb.extractCircleOfTrust(req)

	// Extract proposed outcome
	ctx.ProposedOutcome = cb.extractProposedOutcome(req)

	return ctx
}

// extractUSP infers the Unique Selling Proposition from conversation.
func (cb *ContextBuilder) extractUSP(req *models.ChatRequest) string {
	allText := req.CurrentMessage + " " + cb.concatenateHistory(req.ConversationHistory)

	patterns := []string{
		`(?i)usp[:\s]+([^.!?\n]+)`,
		`(?i)unique(?:\s+selling\s+proposition)?[:\s]+([^.!?\n]+)`,
		`(?i)what\s+makes\s+us\s+different\s+is[:\s]+([^.!?\n]+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(allText)
		if len(matches) > 1 {
			return strings.TrimSpace(matches[1])
		}
	}

	return ""
}

// extractICP infers the Ideal Customer Profile.
func (cb *ContextBuilder) extractICP(req *models.ChatRequest) string {
	allText := req.CurrentMessage + " " + cb.concatenateHistory(req.ConversationHistory)

	patterns := []string{
		`(?i)icp[:\s]+([^.!?\n]+)`,
		`(?i)target(?:\s+audience|\s+customer|\s+market)?[:\s]+([^.!?\n]+)`,
		`(?i)we\s+sell\s+to[:\s]+([^.!?\n]+)`,
		`(?i)my\s+customers\s+are[:\s]+([^.!?\n]+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(allText)
		if len(matches) > 1 {
			return strings.TrimSpace(matches[1])
		}
	}

	return ""
}

// identifyVertical uses heuristics to detect the business vertical.
func (cb *ContextBuilder) identifyVertical(req *models.ChatRequest) string {
	allText := strings.ToLower(req.CurrentMessage + " " + cb.concatenateHistory(req.ConversationHistory))

	verticalKeywords := map[string][]string{
		"supplements":  {"supplement", "vitamin", "nutrition", "protein", "fda", "health"},
		"coaching":     {"coach", "training", "mentorship", "course", "learning", "transformation"},
		"ecommerce":    {"product", "store", "shop", "merchandise", "inventory", "cart"},
		"skincare":     {"skin", "beauty", "cosmetic", "skincare", "serum", "routine"},
		"subscription": {"subscription", "recurring", "membership", "box", "monthly"},
		"nonprofit":    {"nonprofit", "charity", "donation", "cause", "mission", "advocacy"},
		"education":    {"school", "college", "university", "campus", "students", "enroll"},
	}

	for vertical, keywords := range verticalKeywords {
		matchCount := 0
		for _, keyword := range keywords {
			if strings.Contains(allText, keyword) {
				matchCount++
			}
		}
		if matchCount >= 2 {
			return vertical
		}
	}

	return ""
}

// extractCircleOfTrust identifies which Buyer Circle the user is focusing on.
func (cb *ContextBuilder) extractCircleOfTrust(req *models.ChatRequest) string {
	allText := strings.ToLower(req.CurrentMessage + " " + cb.concatenateHistory(req.ConversationHistory))

	// More specific patterns to avoid false positives
	circles := map[string][]string{
		"stranger": {
			"stranger", "cold audience", "cold lead", "new audience", "new prospect",
			"top of funnel", "tofu", "awareness stage", "strangers",
		},
		"follower": {
			"follower", "subscriber", "email subscriber", "newsletter subscriber",
			"warm lead", "engaged audience", "followers",
		},
		"customer": {
			"customer", "buyer", "purchased", "made a purchase", "bought",
			"client", "paid customer", "existing customer", "customers",
		},
		"advocate": {
			"advocate", "loyal customer", "repeat customer", "champion",
			"referral", "brand advocate", "advocates",
		},
	}

	// Require more specific context to avoid false matches
	// Check for Circle of Trust mentions or explicit targeting language
	hasCircleContext := strings.Contains(allText, "circle of trust") ||
		strings.Contains(allText, "buyers' circle") ||
		strings.Contains(allText, "targeting") ||
		strings.Contains(allText, "intended audience") ||
		strings.Contains(allText, "audience is")

	// If no Circle context, be more conservative
	if !hasCircleContext {
		// Only match if it's very explicit
		for circle, keywords := range circles {
			for _, keyword := range keywords {
				// Use word boundaries for more precise matching
				pattern := `\b` + regexp.QuoteMeta(keyword) + `\b`
				matched, _ := regexp.MatchString(pattern, allText)
				if matched {
					// Double-check it's in a relevant context
					if strings.Contains(allText, "circle") || strings.Contains(allText, "audience") ||
						strings.Contains(allText, "target") || strings.Contains(allText, "focus") {
						return circle
					}
				}
			}
		}
		return ""
	}

	// If Circle context exists, do normal matching
	for circle, keywords := range circles {
		for _, keyword := range keywords {
			pattern := `\b` + regexp.QuoteMeta(keyword) + `\b`
			matched, _ := regexp.MatchString(pattern, allText)
			if matched {
				return circle
			}
		}
	}

	return ""
}

// extractProposedOutcome extracts the user's desired outcome.
func (cb *ContextBuilder) extractProposedOutcome(req *models.ChatRequest) string {
	allText := req.CurrentMessage + " " + cb.concatenateHistory(req.ConversationHistory)

	patterns := []string{
		// Explicit goal/outcome mentions (highest priority)
		`(?i)(?:goal|outcome|objective)[:\s]+([^.!?\n]+)`,
		// "My goal is" / "Our goal is" patterns
		`(?i)(?:my|our)\s+goal\s+is\s+([^.!?\n]+)`,
		// "desired outcome" / "desired result"
		`(?i)desired\s+(?:outcome|result)[:\s]+([^.!?\n]+)`,
		// "I want to" / "we want to" patterns
		`(?i)(?:i|we)\s+want\s+to\s+([^.!?\n]+)`,
		// "I need to" / "we need to" patterns
		`(?i)(?:i|we)\s+need\s+to\s+([^.!?\n]+)`,
		// "I'm looking to" / "we're looking to" patterns
		`(?i)(?:i|we)(?:'m|'re)?\s+looking\s+to\s+([^.!?\n]+)`,
		// "I'd like to" / "we'd like to" patterns
		`(?i)(?:i|we)'d\s+like\s+to\s+([^.!?\n]+)`,
		// Specific outcome verbs (avoid generic "get")
		`(?i)(?:achieve|accomplish|obtain|generate|create|build|establish|develop)\s+([^.!?\n]{5,50})`,
	}

	// Common non-outcome phrases to exclude
	excludedPhrases := []string{
		"to", "started", "going", "start", "begin", "beginning",
		"here", "there", "this", "that", "it", "them", "us",
		"more", "less", "better", "worse", "good", "bad",
		"some", "any", "all", "none", "one", "two", "three",
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(allText)
		if len(matches) > 1 {
			extracted := strings.TrimSpace(matches[1])
			// Filter out very short matches
			if len(extracted) < 5 {
				continue
			}
			// Filter out generic/non-outcome phrases
			extractedLower := strings.ToLower(extracted)
			isExcluded := false
			for _, excluded := range excludedPhrases {
				if extractedLower == excluded || strings.HasPrefix(extractedLower, excluded+" ") {
					isExcluded = true
					break
				}
			}
			if isExcluded {
				continue
			}
			// Filter out phrases that are too generic (single common words)
			words := strings.Fields(extractedLower)
			if len(words) == 1 && len(words[0]) < 8 {
				// Single short word is likely not an outcome
				continue
			}
			return extracted
		}
	}

	return ""
}

// hasValidationBeenDone checks if the model has already provided a validation summary
func (cb *ContextBuilder) hasValidationBeenDone(history []instruction.Message) bool {
	// Look for validation indicators in model responses
	validationKeywords := []string{
		"understanding", "summary", "confirm", "correct", "accurate",
		"let me summarize", "to confirm", "based on", "so you",
	}

	for _, msg := range history {
		if msg.Role == "model" || msg.Role == "ai" || msg.Role == "assistant" {
			content := strings.ToLower(msg.Content)
			for _, keyword := range validationKeywords {
				if strings.Contains(content, keyword) {
					// Check if it's a validation summary (mentions USP and/or ICP)
					if strings.Contains(content, "usp") || strings.Contains(content, "unique selling") ||
						strings.Contains(content, "icp") || strings.Contains(content, "ideal customer") ||
						strings.Contains(content, "target audience") {
						return true
					}
				}
			}
		}
	}
	return false
}

// hasCircleConfirmationBeenDone checks if the model has already confirmed the Circle
func (cb *ContextBuilder) hasCircleConfirmationBeenDone(history []instruction.Message) bool {
	// Look for Circle confirmation indicators in model responses
	confirmationKeywords := []string{
		"circle of trust", "buyers' circle", "intended audience",
		"targeting", "focusing on", "audience is",
	}

	for _, msg := range history {
		if msg.Role == "model" || msg.Role == "ai" || msg.Role == "assistant" {
			content := strings.ToLower(msg.Content)
			for _, keyword := range confirmationKeywords {
				if strings.Contains(content, keyword) {
					return true
				}
			}
		}
	}
	return false
}

// hasAnalysisBeenDone checks if the model has already provided the Step 7 analysis
func (cb *ContextBuilder) hasAnalysisBeenDone(history []instruction.Message) bool {
	// Look for analysis indicators in model responses
	analysisKeywords := []string{
		"appropriate", "inappropriate", "aligns", "aligns perfectly",
		"highly appropriate", "well-suited", "matches", "fits",
		"circle of trust", "desired outcome", "analysis",
		"demonstrating", "motivations", "relationship status",
	}

	for _, msg := range history {
		if msg.Role == "model" || msg.Role == "ai" || msg.Role == "assistant" {
			content := strings.ToLower(msg.Content)
			// Check if it contains analysis keywords AND mentions outcome appropriateness
			hasKeywords := false
			for _, keyword := range analysisKeywords {
				if strings.Contains(content, keyword) {
					hasKeywords = true
					break
				}
			}
			if hasKeywords {
				// Verify it's actually an analysis (mentions outcome and circle)
				if (strings.Contains(content, "outcome") || strings.Contains(content, "goal")) &&
					(strings.Contains(content, "circle") || strings.Contains(content, "customer") ||
						strings.Contains(content, "stranger") || strings.Contains(content, "follower") ||
						strings.Contains(content, "advocate")) {
					return true
				}
			}
		}
	}
	return false
}

func (cb *ContextBuilder) concatenateHistory(history []instruction.Message) string {
	var sb strings.Builder
	// no need to enforce maxHistoryLength strictly here, client should trim; but it can be added if needed
	for _, msg := range history {
		sb.WriteString(msg.Content)
		sb.WriteString(" ")
	}
	return sb.String()
}

// DetermineWorkflowStep infers the current workflow step based on conversation state.
func (cb *ContextBuilder) DetermineWorkflowStep(ctx *instruction.UserContext) instruction.WorkflowStep {
	// Determine step based on what information has been gathered
	hasUSP := ctx.ExtractedUSP != ""
	hasICP := ctx.ExtractedICP != ""
	hasCircle := ctx.CurrentCircleOfTrust != ""
	hasOutcome := ctx.ProposedOutcome != ""

	// Count filled fields for simpler logic
	filledFields := 0
	if hasUSP {
		filledFields++
	}
	if hasICP {
		filledFields++
	}
	if hasCircle {
		filledFields++
	}
	if hasOutcome {
		filledFields++
	}

	// Step 1: Introduction - no information yet
	if filledFields == 0 {
		return instruction.StepIntroduction
	}

	// Step 2: Discovery - has either USP or ICP, but not both
	if filledFields == 1 && (hasUSP || hasICP) {
		return instruction.StepDiscovery
	}

	// Steps 3-4: USP and ICP are present
	if hasUSP && hasICP {
		if !hasCircle {
			// Check if validation has already been done by looking at conversation history
			// If model has already provided a validation summary, move to Step 4
			// Otherwise, do Step 3 (Validation) first
			if cb.hasValidationBeenDone(ctx.ConversationHistory) {
				return instruction.StepFrameworkApplication
			}
			// Step 3: Validation - confirming USP/ICP understanding
			return instruction.StepValidation
		}
		// Steps 5-6: Circle is also present
		if hasCircle && !hasOutcome {
			// Check if Circle confirmation has already been done
			// If model has already confirmed Circle, move to Step 6
			// Otherwise, do Step 5 (Circle Confirmation) first
			if cb.hasCircleConfirmationBeenDone(ctx.ConversationHistory) {
				return instruction.StepGoalSetting
			}
			// Step 5: Circle Confirmation - confirming Circle understanding
			return instruction.StepCircleConfirmation
		}
	}

	// Step 7: Analysis - has all 4 fields, analyzing appropriateness
	if filledFields == 4 && hasUSP && hasICP && hasCircle && hasOutcome {
		// Check if analysis has already been done
		if cb.hasAnalysisBeenDone(ctx.ConversationHistory) {
			return instruction.StepExecution
		}
		return instruction.StepAnalysis
	}

	// Default fallback
	return instruction.StepIntroduction
}
