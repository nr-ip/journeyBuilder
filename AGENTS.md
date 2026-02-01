# JourneyBuilder Agent Guidelines

This document provides guidelines for AI coding agents working on JourneyBuilder, a Go-based AI chatbot API for DTC email marketers using Vertex AI (Gemini).

## Build, Lint, and Test Commands

### Building
```bash
# Build the main API server
go build -o journey-builder ./cmd/api

# Build with race detector
go build -race -o journey-builder ./cmd/api
```

### Running
```bash
# Run the server directly from the project root
./journey-builder

# Run with environment variables
PORT=3000 ./journey-builder

# For development with hot-reloading (if air is installed)
# go install github.com/cosmtrek/air@latest
# air
```

### Testing
**Note:** Currently, no `_test.go` files exist. When adding tests, use standard Go testing practices.

```bash
# Run all tests
go test ./...

# Run tests for a specific package
go test ./internal/orchestrator

# Run a specific test function
go test -run TestMyFunction ./internal/orchestrator

# Run tests with race detection and coverage
go test -race -cover ./...
```

### Linting and Formatting
```bash
# Format code (do this before every commit)
go fmt ./...

# Fix imports (do this before every commit)
# go install golang.org/x/tools/cmd/goimports@latest
goimports -w .

# Find common issues
go vet ./...

# Run comprehensive linter (if installed)
# go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
golangci-lint run

# Tidy module dependencies
go mod tidy
```

## Code Style Guidelines

### General
- Follow standard Go naming: `camelCase` for internal variables/functions, `PascalCase` for exported.
- Keep functions small, focused, and under ~50 lines.
- Write clear, descriptive names for variables, functions, and types.
- Add documentation comments to all exported functions and types.
- Use meaningful, wrapped errors. Avoid panicking outside of main initialization.

### Imports
Group imports in the following order, separated by blank lines:
1. Standard library (`context`, `encoding/json`, `fmt`, `net/http`)
2. Third-party packages (`github.com/gorilla/mux`)
3. Internal project packages (`JourneyBuilder/internal/models`)

```go
import (
    "context"
    "encoding/json"

    "github.com/gorilla/mux"

    "JourneyBuilder/internal/models"
)
```

### Error Handling
- Handle all errors. Return errors to the caller instead of logging and returning `nil`.
- Use `fmt.Errorf("...: %w", err)` to wrap errors with context.

```go
func (s *Service) Process(req *Request) error {
    if err := s.validate(req); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }
    // ...
    return nil
}
```

### Types and Structs
- Use meaningful names for structs and interfaces.
- Add `json:"..."` tags for all fields in API-facing structs.
- Use pointer receivers for methods that modify the struct.

```go
type ChatRequest struct {
    CurrentMessage      string    `json:"currentMessage"`
    ConversationHistory []Message `json:"conversationHistory"`
}
```

### Package Organization
- `cmd/`: Application entry points.
- `internal/`: All private application code.
  - `api/`: HTTP handlers and routing.
  - `models/`: Core data structures (request/response types).
  - `services/`: External service integrations (e.g., Gemini).
  - `orchestrator/`: Business logic coordination.
  - `validation/`: Input/output validation.
  - `knowledge/`: Knowledge base management.
- `public/`: Static assets for the frontend.

## Project-Specific Patterns

- **Dependency Injection:** Dependencies (like services and the knowledge base) are initialized in `main.go` and passed to the components that need them.
- **Custom Logger:** Use the logger from `JourneyBuilder/internal/logger` for all logging. It provides structured logging to both console and a file.
- **Environment Variables:** Configuration is managed via environment variables loaded from a `.env` file using `godotenv`. See `main.go` for required variables.

## No Cursor or Copilot Rules Found

No `.cursorrules`, `.cursor/rules/`, or `.github/copilot-instructions.md` files were found in this repository.

