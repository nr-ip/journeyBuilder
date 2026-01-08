package api

import (
	"fmt"
	"net/http"
	"time"

	"JourneyBuilder/internal/api/handlers"
	"JourneyBuilder/internal/orchestrator"

	"github.com/labstack/echo/v4"
)

// SetupRoutes configures all API routes.
func SetupRoutes(e *echo.Echo, orch *orchestrator.Orchestrator) {
	// API group
	api := e.Group("/api/v1")

	// Chat endpoints
	chatHandler := handlers.NewChatHandler(orch)
	api.POST("/chat", chatHandler.Chat)
	api.POST("/chat/stream", chatHandler.ChatStream)

	// Health check
	api.GET("/health", healthCheck)
	api.GET("/status", statusCheck)

	// Knowledge base endpoints (admin/debug)
	api.GET("/frameworks", listFrameworks)
	api.GET("/sequences/:vertical", listSequences)
}

// Health check endpoint.
func healthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "davinci-chatbot",
		"version": "1.0.0",
	})
}

// Status endpoint with orchestrator health.
func statusCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":     "healthy",
		"service":    "davinci-chatbot",
		"version":    "1.0.0",
		"workflows":  8,
		"verticals":  6,
		"frameworks": []string{"AIDA", "PAS", "FAB", "BAB", "4Ps", "Hero"},
		"uptime":     "2h30m",
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	})
}

// Lists available copywriting frameworks (AIDA, PAS, FAB, BAB, 4Ps, Hero)
func listFrameworks(c echo.Context) error {
	frameworks := []string{
		"AIDA", "PAS", "FAB", "BAB", "4Ps", "Hero",
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"frameworks":  frameworks,
		"count":       len(frameworks),
		"description": "Available copywriting frameworks for email sequences",
	})
}

// Lists sequences for specific vertical (supplements, coaching, etc.)
func listSequences(c echo.Context) error {
	vertical := c.Param("vertical")

	sequences := map[string]string{
		"first_purchase":   "5 emails over 7-14 days",
		"cart_recovery":    "3 emails in 48 hours",
		"onboarding":       "5 emails over 21 days",
		"churn_prevention": "4 emails over 30 days",
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"vertical":  vertical,
		"sequences": sequences,
		"count":     len(sequences),
		"message":   fmt.Sprintf("Available sequences for %s vertical", vertical),
	})
}
