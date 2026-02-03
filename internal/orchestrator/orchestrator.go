package orchestrator

import (
	"context"
	"fmt"
	"sync"
	"time"

	"JourneyBuilder/internal/instruction"
	"JourneyBuilder/internal/knowledge"
	"JourneyBuilder/internal/models"
	"JourneyBuilder/internal/services"
	"JourneyBuilder/internal/validation"
)

// Orchestrator coordinates validation, context building, prompt composition, and Gemini AI calls.
type SessionData struct {
	OffTopicStrikeCount int
	LastInteractionTime time.Time
	IsSessionEnded      bool
}

type Orchestrator struct {
	geminiService   *services.GeminiService
	contextBuilder  *ContextBuilder
	kb              *knowledge.KnowledgeBase
	inputValidator  *validation.InputValidator
	outputValidator *validation.OutputValidator

	sessionState map[string]*SessionData
	mu           sync.Mutex // Mutex to protect sessionState
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
		sessionState:    make(map[string]*SessionData),
	}
}

// ProcessChatRequest orchestrates the full non-streaming flow.
func (o *Orchestrator) ProcessChatRequest(
	ctx context.Context,
	req *models.ChatRequest,
	_ bool, // reserved for future flags
) (*models.ChatResponse, error) {
	o.mu.Lock()
	defer o.mu.Unlock()

	// 1. Validate input (prompt injection / security)
	if err := o.inputValidator.ValidateInput(req.CurrentMessage); err != nil {
		// If a security validation error occurs, it's not an off-topic strike but a hard rejection.
		return &models.ChatResponse{
			Message: "I'm sorry, but I cannot fulfill that request as it conflicts with my core operational security protocols.",
			Error:   err.Error(),
		}, fmt.Errorf("security validation failed: %w", err)
	}

	session, exists := o.sessionState[req.SessionID]
	if !exists {
		session = &SessionData{
			OffTopicStrikeCount: 0,
			LastInteractionTime: time.Now(),
			IsSessionEnded:      false,
		}
		o.sessionState[req.SessionID] = session
	} else {
		// Check for inactivity (e.g., 30 minutes)
		if time.Since(session.LastInteractionTime) > 30*time.Minute {
			session.OffTopicStrikeCount = 0 // Reset strikes on inactivity
			session.IsSessionEnded = false  // Allow new session after inactivity
		}
	}

	if session.IsSessionEnded {
		return &models.ChatResponse{
			Message:        "This session has ended due to repeated off-topic inputs. Please start a new chat if you need assistance with email journeys.",
			IsSessionEnded: true,
		}, nil
	}

	// 2. Build stateless context
	userCtx := o.contextBuilder.BuildContext(req)

	// Determine if the input is off-topic based on extracted context
	// An input is considered off-topic if none of the core context fields are extracted.
	onTopic := userCtx.ExtractedUSP != "" ||
		userCtx.ExtractedICP != "" ||
		userCtx.IdentifiedVertical != "" ||
		userCtx.CurrentCircleOfTrust != "" ||
		userCtx.ProposedOutcome != ""

	if !onTopic {
		session.OffTopicStrikeCount++
		session.LastInteractionTime = time.Now()

		var responseMessage string
		sessionEnded := false

		switch session.OffTopicStrikeCount {
		case 1:
			responseMessage = "I'm here to help you build and manage email journeys. Can you please ask questions related to that topic?"
		case 2:
			responseMessage = "It seems we're discussing topics outside of email journey support. To keep us on track, please focus on email journey-related questions."
		case 3:
			responseMessage = "Thank you for your input. As this is the third consecutive request outside the scope of email journey support, this session will now end. Please start a new chat if you need assistance with email journeys."
			sessionEnded = true
		default:
			// Should not happen if logic is correct, but as a fallback
			responseMessage = "This session has ended due to repeated off-topic inputs. Please start a new chat if you need assistance with email journeys."
			sessionEnded = true
		}

		session.IsSessionEnded = sessionEnded
		// Store updated session data
		o.sessionState[req.SessionID] = session

		return &models.ChatResponse{
			Message:        responseMessage,
			IsSessionEnded: sessionEnded,
		}, nil
	} else {
		// If on-topic, reset strike counter
		session.OffTopicStrikeCount = 0
		session.LastInteractionTime = time.Now()
		o.sessionState[req.SessionID] = session
	}

	// If not off-topic and not session ended, proceed with normal processing.

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
			MaxEmailLength:   750,
			ReadabilityLevel: "Grade6",
		},
	}

	composedPrompt := composerCfg.ComposeInstructions()

	// 6. Build Gemini AI request
	// Convert models.Message to services.Message
	convHistory := make([]services.Message, len(req.ConversationHistory))
	for i, msg := range req.ConversationHistory {
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
		MaxTokens:           3000,
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
			out <- fmt.Sprintf("event: error\ndata: Error processing your request: %s\n\n", err.Error())
			return
		}
		if resp.IsSessionEnded {
			out <- fmt.Sprintf("event: message\ndata: %s\n\n", resp.Message)
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
