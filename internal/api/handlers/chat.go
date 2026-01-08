package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"JourneyBuilder/internal/models"
	"JourneyBuilder/internal/orchestrator"
)

// ChatHandler wires HTTP layer to the orchestrator.
type ChatHandler struct {
	orch *orchestrator.Orchestrator
}

func NewChatHandler(orch *orchestrator.Orchestrator) *ChatHandler {
	return &ChatHandler{orch: orch}
}

// Health returns a simple health check payload.
func (h *ChatHandler) Health(c echo.Context) error {
	resp := models.HealthCheckResponse{
		Status:  "ok",
		Message: "Da Vinci API is running",
	}
	return c.JSON(http.StatusOK, resp)
}

// Chat handles non-streaming chat requests: POST /api/chat
func (h *ChatHandler) Chat(c echo.Context) error {
	var req models.ChatRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error": "invalid request body",
		})
	}

	resp, err := h.orch.ProcessChatRequest(c.Request().Context(), &req, false)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error":   "failed to process request",
			"details": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, resp)
}

// ChatStream handles streaming chat: POST /api/chat/stream
// For simplicity this example uses Server-Sent Events (SSE) style text streaming.
func (h *ChatHandler) ChatStream(c echo.Context) error {
	// You can choose a different content type if you implement true SSE.
	c.Response().Header().Set(echo.HeaderContentType, "text/event-stream")
	c.Response().WriteHeader(http.StatusOK)
	c.Response().Flush()

	var req models.ChatRequest
	if err := c.Bind(&req); err != nil {
		_, _ = c.Response().Write([]byte("event: error\ndata: invalid request body\n\n"))
		c.Response().Flush()
		return nil
	}

	stream, err := h.orch.ProcessChatRequestStream(c.Request().Context(), &req)
	if err != nil {
		_, _ = c.Response().Write([]byte("event: error\ndata: failed to start stream\n\n"))
		c.Response().Flush()
		return nil
	}

	for chunk := range stream {
		// Each chunk is just text; you can adapt the format as needed.
		_, _ = c.Response().Write([]byte("event: message\n"))
		_, _ = c.Response().Write([]byte("data: " + chunk + "\n\n"))
		c.Response().Flush()
	}

	return nil
}
