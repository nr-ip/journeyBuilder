package models

// ChatRequest is the payload received from the frontend.
type ChatRequest struct {
	SessionID           string         `json:"sessionId"`
	CurrentMessage      string         `json:"currentMessage"`
	ConversationHistory []Message      `json:"conversationHistory"`
	BaseSystemPrompt    string         `json:"baseSystemPrompt"`       // optional override
	UserMetadata        map[string]any `json:"userMetadata,omitempty"` // optional extra context
}

// ChatResponse is the structured response returned to the frontend.
type ChatResponse struct {
	Message            string `json:"message"`
	WorkflowStep       int    `json:"workflowStep"`
	ExtractedUSP       string `json:"extractedUSP,omitempty"`
	ExtractedICP       string `json:"extractedICP,omitempty"`
	IdentifiedVertical string `json:"identifiedVertical,omitempty"`
	CurrentCircle      string `json:"currentCircle,omitempty"`
	ProposedOutcome    string `json:"proposedOutcome,omitempty"`
	IsSessionEnded     bool   `json:"isSessionEnded,omitempty"`
	Error              string `json:"error,omitempty"`
}

// HealthCheckResponse is used for /health endpoint.
type HealthCheckResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// Message represents a message in the conversation history.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
