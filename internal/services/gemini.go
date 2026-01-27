package services

import (
	"JourneyBuilder/internal/logger"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
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

	if err := godotenv.Load(); err != nil {
		log.Println("Info: .env file not found, using system environment variables")
	} else {
		log.Println("✓ Loaded .env file")
	}
	projectID := os.Getenv("GCP_PROJECT_ID")
	if projectID == "" {
		return nil, fmt.Errorf("GCP_PROJECT_ID is not set")
	}
	location := os.Getenv("GCP_REGION")
	if location == "" {
		location = "us-central1"
	}
	model := os.Getenv("GEMINI_MODEL")
	if model == "" {
		model = "gemini-2.5-flash"
	}

	ctx := context.Background()

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		Project:  projectID,
		Location: location,
		Backend:  genai.BackendVertexAI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}
	logger.Println("✓ Gemini client created")

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
		logger.Printf("📝 GENERATED SYSTEM PROMPT:\n%s\n", req.SystemPrompt)
		config.SystemInstruction = &genai.Content{
			Parts: []*genai.Part{
				{Text: req.SystemPrompt},
			},
		}
	}

	// Add conversation history
	for _, msg := range req.ConversationHistory {
		// Convert "ai" role to "model" as required by Gemini API
		role := msg.Role
		if role == "ai" || role == "assistant" {
			role = "model"
		}
		// Ensure role is either "user" or "model"
		if role != "user" && role != "model" {
			role = "user" // Default to "user" if invalid
		}
		contents = append(contents, &genai.Content{
			Role:  role,
			Parts: []*genai.Part{{Text: msg.Content}},
		})
	}

	logger.Printf("👤 USER MESSAGE:\n%s\n", req.UserMessage)
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

	// Generate content with retry logic for transient errors
	var resp *genai.GenerateContentResponse
	var err error
	maxRetries := 3
	retryDelay := time.Second

	for attempt := 0; attempt < maxRetries; attempt++ {
		resp, err = c.client.Models.GenerateContent(ctx, c.model, contents, config)
		if err == nil {
			break
		}

		// Check if it's a retryable error (503, 429, or temporary errors)
		errStr := strings.ToLower(err.Error())
		isRetryable := false
		if errStr != "" {
			// Check for common retryable error patterns
			if strings.Contains(errStr, "503") || strings.Contains(errStr, "overloaded") ||
				strings.Contains(errStr, "429") || strings.Contains(errStr, "rate limit") ||
				strings.Contains(errStr, "unavailable") || strings.Contains(errStr, "resource_exhausted") {
				isRetryable = true
			}
		}

		if !isRetryable || attempt == maxRetries-1 {
			return &Response{Error: fmt.Sprintf("Failed to generate the content: %v", err)}, err
		}

		// Wait before retrying with exponential backoff
		time.Sleep(retryDelay * time.Duration(attempt+1))
	}

	if err != nil {
		return &Response{Error: fmt.Sprintf("Failed to generate the content after %d attempts: %v", maxRetries, err)}, err
	}

	// Extract text from response
	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
		return &Response{Text: ""}, nil
	}

	responseText := resp.Candidates[0].Content.Parts[0].Text
	logger.Printf("🤖 MODEL RESPONSE:\n%s\n", responseText)
	return &Response{Text: responseText}, nil
}
