package services

import (
	"context"
	"fmt"
	"os"

	"google.golang.org/genai"
)

// Client is a wrapper around Gemini GenerativeModel.
type GeminiService struct {
	client *genai.Client
	model  string
}

// Message represents a chat message with a role and content.
type Message struct {
	Role    string
	Content string
}

// RequestBuilder configures the Gemini AI request.
type RequestBuilder struct {
	SystemPrompt        string
	UserMessage         string
	ConversationHistory []Message
	Temperature         float32
	MaxTokens           int
}

// Response contains the model output.
type Response struct {
	Text  string
	Error string
}

func NewGeminiService() (*GeminiService, error) {
	// Try GEMINI_API_KEY first (for direct API access)
	apiKey := os.Getenv("GEMINI_API_KEY")

	// If not set, check if GOOGLE_APPLICATION_CREDENTIALS is set (for Vertex AI)
	if apiKey == "" {
		credsPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
		if credsPath != "" {
			return nil, fmt.Errorf("GOOGLE_APPLICATION_CREDENTIALS is set but GEMINI_API_KEY is required for genai client. Please set GEMINI_API_KEY environment variable")
		}
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable not set. Please set it in your .env file or export it: export GEMINI_API_KEY=your_api_key")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}
	model := os.Getenv("GEMINI_MODEL")
	if model == "" {
		model = "gemini-2.5-flash"
	}

	return &GeminiService{
		client: client,
		model:  model,
	}, nil
}

// Close closes the underlying genai client connection.
// Note: genai.Client may not have a Close method, so this is a no-op for now.
func (c *GeminiService) Close() error {
	// genai.Client doesn't require explicit closing in current API
	// Resources are managed by the library
	return nil
}

// SendRequest sends a request to Gemini.
func (c *GeminiService) SendRequest(ctx context.Context, req *RequestBuilder) (*Response, error) {
	// Build content parts
	var contents []*genai.Content

	config := &genai.GenerateContentConfig{}

	if req.SystemPrompt != "" {
		// Correct way: Initialize the struct pointer directly:
		// "System Prompt" (the instructions that tell the AI how to behave), you shouldn't append it to the contents slice with a user role.
		// Instead, you pass it in the Config object.
		config.SystemInstruction = &genai.Content{
			Parts: []*genai.Part{
				{Text: req.SystemPrompt},
			},
		}
	}

	// Add conversation history
	for _, msg := range req.ConversationHistory {
		contents = append(contents, &genai.Content{
			Role:  msg.Role,
			Parts: []*genai.Part{{Text: msg.Content}},
		})
	}

	// Current user message
	contents = append(contents, &genai.Content{
		Role:  "user",
		Parts: []*genai.Part{{Text: req.UserMessage}},
	})

	if req.Temperature != 0 {
		t := float32(req.Temperature)
		config.Temperature = &t
	}
	if req.MaxTokens != 0 {
		m := int32(req.MaxTokens)
		config.MaxOutputTokens = m
	}

	// Generate content
	resp, err := c.client.Models.GenerateContent(ctx, c.model, contents, config)
	if err != nil {
		return &Response{Error: fmt.Sprintf("Failed to generate the content: %v", err)}, err
	}

	// Extract text from response
	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
		return &Response{Text: ""}, nil
	}

	return &Response{Text: resp.Candidates[0].Content.Parts[0].Text}, nil
}
