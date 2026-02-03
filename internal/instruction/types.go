package instruction

import (
	"JourneyBuilder/internal/models"
)

// UserContext holds extracted context from stateless conversation analysis.
type UserContext struct {
	ConversationHistory  []models.Message
	ExtractedUSP         string
	ExtractedICP         string
	IdentifiedVertical   string
	CurrentCircleOfTrust string
	ProposedOutcome      string
}

// WorkflowStep represents the 8-step conversation workflow.
type WorkflowStep int

const (
	StepIntroduction WorkflowStep = iota
	StepDiscovery
	StepValidation
	StepFrameworkApplication
	StepCircleConfirmation
	StepGoalSetting
	StepAnalysis
	StepExecution
)

// OutputFormat controls the response structure.
type OutputFormat struct {
	Type             string   `json:"type"`
	IncludeTable     bool     `json:"includeTable"`
	TableColumns     []string `json:"tableColumns"`
	MaxEmailLength   int      `json:"maxEmailLength"`
	ReadabilityLevel string   `json:"readabilityLevel"`
}

// ComposerConfig defines inputs for prompt composition.
type ComposerConfig struct {
	BaseSystemPrompt string
	WorkflowStep     WorkflowStep
	UserContext      UserContext
	VerticalType     string
	KnowledgeContext string
	OutputFormat     OutputFormat
}
