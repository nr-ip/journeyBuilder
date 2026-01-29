package handlers

import (
	"encoding/json"
	"net/http"

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
func (h *ChatHandler) Health(w http.ResponseWriter, r *http.Request) {
	resp := models.HealthCheckResponse{
		Status:  "ok",
		Message: "Da Vinci API is running",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// Chat handles non-streaming chat requests: POST /api/chat
func (h *ChatHandler) Chat(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req models.ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "invalid request body",
		})
		return
	}

	resp, err := h.orch.ProcessChatRequest(r.Context(), &req, false)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "failed to process request",
			"details": err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// ChatStream handles streaming chat: POST /api/chat/stream
// For simplicity this example uses Server-Sent Events (SSE) style text streaming.
func (h *ChatHandler) ChatStream(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// You can choose a different content type if you implement true SSE.
	w.Header().Set("Content-Type", "text/event-stream")
	w.WriteHeader(http.StatusOK)
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}

	var req models.ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_, _ = w.Write([]byte("event: error\ndata: invalid request body\n\n"))
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
		return
	}

	stream, err := h.orch.ProcessChatRequestStream(r.Context(), &req)
	if err != nil {
		_, _ = w.Write([]byte("event: error\ndata: failed to start stream\n\n"))
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
		return
	}

	flusher, hasFlusher := w.(http.Flusher)
	for chunk := range stream {
		// Each chunk is just text; you can adapt the format as needed.
		_, _ = w.Write([]byte("event: message\n"))
		_, _ = w.Write([]byte("data: " + chunk + "\n\n"))
		if hasFlusher {
			flusher.Flush()
		}
	}
}
