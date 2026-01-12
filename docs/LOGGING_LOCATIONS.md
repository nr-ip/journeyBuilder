# Logging Locations in JourneyBuilder

This document lists all locations where the application writes logs using Go's standard `log` package.

## Overview

The application uses Go's standard `log` package, which writes to **stderr (standard error)** by default. All logs appear in the terminal/console where the application is running.

---

## Logging Locations

### 1. `cmd/api/main.go` - Main Application Entry Point

**File**: `cmd/api/main.go`

**Logging Points**:

| Line | Function | Log Message | Type |
|------|----------|-------------|------|
| 25 | `main()` | `"Info: .env file not found, using system environment variables"` | Info |
| 27 | `main()` | `"✓ Loaded .env file"` | Info |
| 38 | `main()` | `"Initializing AI services..."` | Info |
| 41 | `main()` | `"Failed to initialize Gemini service: %v"` | Fatal (exits) |
| 51 | `main()` | `"Failed to initialize knowledge base: %v"` | Fatal (exits) |
| 90 | `main()` | `"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"` | Info |
| 91 | `main()` | `"🚀 Server starting on port %s"` | Info |
| 92 | `main()` | `"📱 Open http://localhost:%s in your browser"` | Info |
| 93 | `main()` | `"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"` | Info |
| 95 | `main()` | Server error (from `http.ListenAndServe`) | Fatal (exits) |
| 148 | `setupGracefulShutdown()` | `"\n🛑 Shutting down gracefully..."` | Info |
| 150 | `setupGracefulShutdown()` | `"Error closing Gemini service: %v"` | Error |
| 152 | `setupGracefulShutdown()` | `"✓ Cleanup complete"` | Info |

**Purpose**: 
- Application startup/shutdown logging
- Service initialization status
- Error handling (fatal errors cause application exit)

---

### 2. `internal/services/gemini.go` - Gemini AI Service

**File**: `internal/services/gemini.go`

**Logging Points**:

| Line | Function | Log Message | Type |
|------|----------|-------------|------|
| 45 | `NewGeminiService()` | `"Info: .env file not found, using system environment variables"` | Info |
| 47 | `NewGeminiService()` | `"✓ Loaded .env file"` | Info |
| 72 | `NewGeminiService()` | `"✓ Gemini client created"` | Info |

**Purpose**:
- Gemini service initialization
- Environment variable loading status
- Client creation confirmation

---

### 3. `internal/api/handlers/init.go` - Handler Initialization

**File**: `internal/api/handlers/init.go`

**Logging Points**:

| Line | Function | Log Message | Type |
|------|----------|-------------|------|
| 16 | `InitializeGeminiService()` | `"✓ Gemini AI service initialized"` | Info |

**Purpose**:
- Handler initialization confirmation

---

### 4. `internal/api/handlers/middleware.go` - HTTP Middleware

**File**: `internal/api/handlers/middleware.go`

**Logging Points**:

| Line | Function | Log Message | Type |
|------|----------|-------------|------|
| 26 | `StandardLogger.Infof()` | Uses `log.Printf()` internally | Info |
| 64 | `RequestLogger()` | Request details: `"%s %s uid=%s from=%s"` | Info |
| 80 | `RequestLogger()` | Response details: `"%s %s uid=%s status=%d latency=%v"` | Info |

**Purpose**:
- HTTP request/response logging
- Request ID tracking
- Performance metrics (latency)
- Client IP logging

**Note**: The middleware uses a `Logger` interface, but the default `StandardLogger` implementation uses `log.Printf()`.

---

## Log Output Destination

### Default Behavior
- **Destination**: `os.Stderr` (standard error)
- **Format**: Plain text with timestamps (if `log.SetFlags()` is configured)
- **No file logging**: Logs are not written to files by default

### Where Logs Appear

1. **Terminal/Console**: When running directly (`go run` or `./binary`)
   ```
   $ go run cmd/api/main.go
   Info: .env file not found, using system environment variables
   Initializing AI services...
   ✓ Gemini client created
   🚀 Server starting on port 8080
   ```

2. **Docker Logs**: When running in a container
   ```bash
   docker logs <container-name>
   ```

3. **System Logs**: When running as a systemd service
   ```bash
   journalctl -u journey-builder
   ```

4. **Process Manager**: When using process managers like PM2, Supervisor, etc.

---

## Log Types Used

### `log.Println()`
- Simple message logging
- Adds newline automatically
- Used for: Info messages, status updates

### `log.Printf()`
- Formatted message logging
- Supports format strings (`%s`, `%v`, `%d`, etc.)
- Used for: Dynamic messages with variables

### `log.Fatalf()` / `log.Fatal()`
- Fatal error logging
- **Exits the application** after logging
- Used for: Critical errors that prevent startup

---

## Example Log Output

When the application starts, you'll see logs like:

```
Info: .env file not found, using system environment variables
Initializing AI services...
✓ Gemini client created
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
🚀 Server starting on port 8080
📱 Open http://localhost:8080 in your browser
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

When a request comes in (if middleware is enabled):
```
POST /api/chat uid=abc123 from=127.0.0.1
POST /api/chat uid=abc123 status=200 latency=1.2s
```

---

## Summary

**Total Logging Locations**: 4 files
- `cmd/api/main.go`: 12 log statements
- `internal/services/gemini.go`: 3 log statements
- `internal/api/handlers/init.go`: 1 log statement
- `internal/api/handlers/middleware.go`: 2 log statements (via interface)

**All logs write to**: `os.Stderr` (standard error stream)

**No file logging configured**: Logs are not written to disk files

---

*Last updated: Based on current codebase structure*
