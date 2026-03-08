### Executive Summary

- **Risk Score**: [🟡 MEDIUM]
- **Summary**: The PR significantly improves the server's graceful shutdown mechanism, making it more robust and aligned with standard Go practices. However, it introduces a regression by omitting the cleanup call for the Gemini service, which could lead to resource leaks.

### Findings

#### 1. Resource Leak on Shutdown

**File**: `cmd/api/main.go`

**Severity**: Medium

**Description**: The new graceful shutdown implementation correctly closes the HTTP server, but it no longer calls `geminiService.Close()`. The previous implementation handled this cleanup. This omission will result in the application terminating without properly releasing resources used by the Gemini client, potentially leading to leaked connections or other resources.

**Suggestion**: The `geminiService.Close()` call should be added to the shutdown signal handler to ensure all resources are cleaned up correctly.

In `cmd/api/main.go`:
```suggestion
		// Trigger graceful shutdown
		logger.Println("
🛑 Shutting down gracefully...")
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Fatalf("Server shutdown failed: %v", err)
		}
		if err := geminiService.Close(); err != nil {
			logger.Printf("Error closing Gemini service: %v", err)
		}
		logger.Println("✓ Cleanup complete")
		serverStopCtx()
```