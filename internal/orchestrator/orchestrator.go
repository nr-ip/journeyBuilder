package orchestrator

import (
	"context"

	"JourneyBuilder/internal/instruction"
	"JourneyBuilder/internal/knowledge"
	"JourneyBuilder/internal/models"
	"JourneyBuilder/internal/services"
	"JourneyBuilder/internal/validation"
)

// Orchestrator coordinates validation, context building, prompt composition, and Gemini AI calls.
type Orchestrator struct {
	geminiService   *services.GeminiService
	contextBuilder  *ContextBuilder
	kb              *knowledge.KnowledgeBase
	inputValidator  *validation.InputValidator
	outputValidator *validation.OutputValidator
}

// NewOrchestrator wires all core services together.
func NewOrchestrator(
	geminiService *services.GeminiService,
	kb *knowledge.KnowledgeBase,
	inputValidator *validation.InputValidator,
	outputValidator *validation.OutputValidator,
) *Orchestrator {
	return &Orchestrator{
		geminiService:   geminiService,
		contextBuilder:  NewContextBuilder(20),
		kb:              kb,
		inputValidator:  inputValidator,
		outputValidator: outputValidator,
	}
}

// ProcessChatRequest orchestrates the full non-streaming flow.
func (o *Orchestrator) ProcessChatRequest(
	ctx context.Context,
	req *models.ChatRequest,
	_ bool, // reserved for future flags
) (*models.ChatResponse, error) {
	// 1. Validate input (prompt injection / security)
	if err := o.inputValidator.ValidateInput(req.CurrentMessage); err != nil {
		return &models.ChatResponse{
			Message: "I'm sorry, but I cannot fulfill that request as it conflicts with my core operational security protocols.",
			Error:   err.Error(),
		}, err
	}

	// 2. Build stateless context
	userCtx := o.contextBuilder.BuildContext(req)

	// 3. Determine workflow step
	currentStep := o.contextBuilder.DetermineWorkflowStep(userCtx)

	// 4. Extract optimized knowledge from KB
	// Convert WorkflowStep to string for knowledge base lookup
	stepStr := workflowStepToString(currentStep)
	kbContext := o.kb.ExtractRelevantContext(userCtx.ProposedOutcome, userCtx.IdentifiedVertical, stepStr)

	// 5. Compose modular instructions
	composerCfg := &instruction.ComposerConfig{
		BaseSystemPrompt: req.BaseSystemPrompt,
		WorkflowStep:     currentStep,
		UserContext:      *userCtx,
		VerticalType:     userCtx.IdentifiedVertical,
		KnowledgeContext: kbContext,
		OutputFormat: instruction.OutputFormat{
			Type:             "text",
			IncludeTable:     shouldIncludeTable(currentStep),
			TableColumns:     []string{"Email #", "Subject Line", "Day Delay"},
			MaxEmailLength:   220,
			ReadabilityLevel: "Grade6",
		},
	}

	composedPrompt := composerCfg.ComposeInstructions()

	// 6. Build Gemini AI request
	// Convert instruction.Message to services.Message
	convHistory := make([]services.Message, len(userCtx.ConversationHistory))
	for i, msg := range userCtx.ConversationHistory {
		convHistory[i] = services.Message{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	geminiReq := &services.RequestBuilder{
		SystemPrompt:        composedPrompt,
		UserMessage:         req.CurrentMessage,
		ConversationHistory: convHistory,
		Temperature:         0.7,
		MaxTokens:           1500,
	}

	// 7. Call Gemini AI
	resp, err := o.geminiService.SendRequest(ctx, geminiReq)
	if err != nil {
		return &models.ChatResponse{
			Message: "Error processing your request. Please try again.",
			Error:   err.Error(),
		}, err
	}

	// 8. Validate output (spam/compliance)
	_ = o.outputValidator.ValidateResponse(resp.Text, currentStep)

	// 9. Return structured response
	return &models.ChatResponse{
		Message:            resp.Text,
		WorkflowStep:       int(currentStep),
		ExtractedUSP:       userCtx.ExtractedUSP,
		ExtractedICP:       userCtx.ExtractedICP,
		IdentifiedVertical: userCtx.IdentifiedVertical,
		CurrentCircle:      userCtx.CurrentCircleOfTrust,
		ProposedOutcome:    userCtx.ProposedOutcome,
	}, nil
}

// ProcessChatRequestStream is a placeholder streaming API.
// For now, it just returns a single chunk channel from the non-streaming call.
func (o *Orchestrator) ProcessChatRequestStream(
	ctx context.Context,
	req *models.ChatRequest,
) (<-chan string, error) {
	out := make(chan string, 1)

	go func() {
		defer close(out)
		resp, err := o.ProcessChatRequest(ctx, req, true)
		if err != nil {
			out <- "Error processing your request. Please try again."
			return
		}
		out <- resp.Message
	}()

	return out, nil
}

func shouldIncludeTable(step instruction.WorkflowStep) bool {
	return step == instruction.StepExecution
}

// workflowStepToString converts a WorkflowStep to its string representation for knowledge base lookups
func workflowStepToString(step instruction.WorkflowStep) string {
	stepMap := map[instruction.WorkflowStep]string{
		instruction.StepIntroduction:         "StepIntroduction",
		instruction.StepDiscovery:            "StepDiscovery",
		instruction.StepValidation:           "StepValidation",
		instruction.StepFrameworkApplication: "StepFrameworkApplication",
		instruction.StepCircleConfirmation:   "StepCircleConfirmation",
		instruction.StepGoalSetting:          "StepGoalSetting",
		instruction.StepAnalysis:             "StepAnalysis",
		instruction.StepExecution:            "StepExecution",
	}
	if str, ok := stepMap[step]; ok {
		return str
	}
	return "StepIntroduction" // default
}
