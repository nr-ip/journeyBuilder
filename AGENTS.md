# JourneyBuilder Agent Guidelines

This document provides guidelines for AI coding agents working on JourneyBuilder, a Go-based AI chatbot API for DTC email marketers using Vertex AI (Gemini).

## Build, Lint, and Test Commands

### Building
```bash
# Build the main API server
go build -o journey-builder ./cmd/api

# Build with race detector
go build -race -o journey-builder ./cmd/api

# Clean build artifacts
go clean
```

### Running
```bash
# Run the server directly
./journey-builder

# Run with environment variables
PORT=3000 ./journey-builder

# Development mode (hot reload with air)
# go install github.com/cosmtrek/air@latest
# air
```

### Testing
**Note:** Currently no test files exist. When adding tests:

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run a specific test function
go test -run TestFunctionName ./path/to/package

# Run tests with race detection
go test -race ./...

# Generate coverage profile
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Linting and Formatting
```bash
# Format code
go fmt ./...

# Check for issues
go vet ./...

# Install and run golangci-lint
# go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
golangci-lint run

# Fix imports (install goimports)
# go install golang.org/x/tools/cmd/goimports@latest
goimports -w .

# Clean dependencies
go mod tidy
```

## Code Style Guidelines

### General Go Conventions
- Follow standard Go naming: camelCase for variables/functions, PascalCase for exported
- Use `gofmt` and `goimports` for formatting
- Write clear, descriptive names
- Add documentation comments for exported functions/types
- Keep functions small and focused
- Use meaningful error messages

### Imports
- Group: internal, blank line, standard library, blank line, third-party
- Full paths for internal packages
- Remove unused imports

```go
import (
    "JourneyBuilder/internal/models"

    "context"
    "encoding/json"

    "github.com/gorilla/mux"
)
```

### Error Handling
- Check and handle errors appropriately
- Return errors to caller (no panicking except fatal issues)
- Use custom error types
- Log with context

```go
func (s *Service) ProcessRequest(req *Request) error {
    if err := s.validateRequest(req); err != nil {
        return fmt.Errorf("invalid request: %w", err)
    }

    result, err := s.process(req)
    if err != nil {
        s.logger.Printf("Failed to process request %v: %v", req.ID, err)
        return err
    }

    return nil
}
```

### Types and Structs
- Meaningful names
- JSON tags for API structs
- Pointer receivers for modifying methods
- Interfaces for dependencies

```go
type ChatRequest struct {
    CurrentMessage      string                `json:"currentMessage"`
    ConversationHistory []Message             `json:"conversationHistory"`
    BaseSystemPrompt    string                `json:"baseSystemPrompt"`
    UserMetadata        map[string]any        `json:"userMetadata,omitempty"`
}
```

### Functions and Methods
- Under 50 lines
- Descriptive names
- Early returns for errors
- Use context for timeouts

```go
func (o *Orchestrator) ProcessChatRequest(
    ctx context.Context,
    req *models.ChatRequest,
    stream bool,
) (*models.ChatResponse, error) {
    if err := o.inputValidator.ValidateInput(req.CurrentMessage); err != nil {
        return &models.ChatResponse{
            Message: "Security protocol violation detected.",
            Error:   err.Error(),
        }, err
    }

    // Core processing
    // ...
}
```

### Constants and Configuration
- Constants for magic numbers
- Group related constants
- Environment variables for config
- Validate required env vars at startup

```go
const (
    DefaultPort     = "8080"
    MaxMessageLength = 10000
    RequestTimeout   = 30 * time.Second
)

func initConfig() error {
    required := []string{"GCP_PROJECT_ID", "GCP_REGION"}
    for _, env := range required {
        if os.Getenv(env) == "" {
            return fmt.Errorf("required environment variable %s not set", env)
        }
    }
    return nil
}
```

### Logging
- Structured logging with levels
- Include context
- Avoid sensitive data
- Use internal logger package

```go
logger.Printf("Processing chat request for vertical: %s", vertical)
```

### Security Best Practices
- Validate all inputs
- Use validation for prompt injection
- Never log sensitive data
- Proper CORS
- HTTPS in production

### Package Organization
- Focused packages
- Internal for private code
- Export only necessary
- Directory structure:
  - `cmd/` - Entry points
  - `internal/` - Private code
  - `internal/models/` - Data structures
  - `internal/services/` - Integrations
  - `internal/validation/` - Validation

### Dependencies
- Minimal and maintained
- Use go mod
- Run `go mod tidy` after changes
- Pin versions

### Performance Considerations
- Efficient structures (maps for lookups)
- Caching (LRU used)
- Profile critical code
- Goroutines for concurrency
- Context for timeouts

### Testing Guidelines (When Adding Tests)
- Unit tests for public functions
- Table-driven tests
- Mock dependencies
- Test errors and edge cases
- >80% coverage

```go
func TestOrchestrator_ProcessChatRequest(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {
            name:    "valid request",
            input:   "I sell supplements",
            expected: "What's your USP?",
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

## Project-Specific Patterns

### AI Service Integration
- Use services package for AI calls
- Proper error handling
- Cache responses
- Validate AI responses

### Workflow Management
- 8-step workflow in orchestrator
- Enums for steps
- Stateless context
- Validate transitions

### Knowledge Base Usage
- Extract relevant context
- Compressed for tokens
- Cache knowledge
- Update files carefully

## Development Workflow

1. **Before coding**: `go mod tidy` and `go fmt ./...`
2. **During development**: `go build` frequently
3. **Before commit**: Format, vet, lint
4. **Testing**: Add tests, run existing
5. **Security**: Validate inputs, no logging sensitive data

## File Organization

- Related functionality in same package
- Descriptive filenames (e.g., `chat_handler.go`)
- Group handlers, models, services
- Document packages

## Common Patterns to Follow

- Dependency injection
- Interfaces for mocking
- Resource cleanup with defer
- Context propagation
- Error wrapping

## No Cursor or Copilot Rules Found

No `.cursorrules` or `.cursor/rules/` directory found. No `.github/copilot-instructions.md` found.

This document should be updated as the codebase evolves.