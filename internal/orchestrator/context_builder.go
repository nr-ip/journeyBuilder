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

	circles := map[string][]string{
		"stranger": {"stranger", "cold", "new audience", "prospect", "top of funnel"},
		"follower": {"follower", "subscriber", "lead", "engaged", "warm"},
		"customer": {"customer", "buyer", "purchase", "client", "paid"},
		"advocate": {"advocate", "loyal", "repeat", "champion", "referral"},
	}

	for circle, keywords := range circles {
		for _, keyword := range keywords {
			if strings.Contains(allText, keyword) {
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
		`(?i)goal[:\s]+([^.!?\n]+)`,
		`(?i)outcome[:\s]+([^.!?\n]+)`,
		`(?i)i\s+want\s+to\s+([^.!?\n]+)`,
		`(?i)we\s+want\s+to\s+([^.!?\n]+)`,
		`(?i)objective[:\s]+([^.!?\n]+)`,
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
	filledFields := 0

	if ctx.ExtractedUSP != "" {
		filledFields++
	}
	if ctx.ExtractedICP != "" {
		filledFields++
	}
	if ctx.CurrentCircleOfTrust != "" {
		filledFields++
	}
	if ctx.ProposedOutcome != "" {
		filledFields++
	}

	switch filledFields {
	case 0:
		return instruction.StepIntroduction
	case 1:
		return instruction.StepDiscovery
	case 2:
		return instruction.StepValidation
	case 3:
		return instruction.StepFrameworkApplication
	case 4:
		return instruction.StepAnalysis
	default:
		return instruction.StepExecution
	}
}
