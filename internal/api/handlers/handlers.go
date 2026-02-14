package handlers

import (
	"encoding/json"
	"net/http"

	"JourneyBuilder/internal/logger"
	"JourneyBuilder/internal/models"
	"JourneyBuilder/internal/orchestrator"
)

// APIHandler holds dependencies for handlers.
type APIHandler struct {
	Orchestrator *orchestrator.Orchestrator
}

// NewAPIHandler creates a new APIHandler with the given orchestrator.
func NewAPIHandler(orch *orchestrator.Orchestrator) *APIHandler {
	return &APIHandler{Orchestrator: orch}
}

// HandleChat handles chat requests using the injected orchestrator.
func (h *APIHandler) HandleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if h.Orchestrator == nil {
		http.Error(w, "Orchestrator not initialized", http.StatusInternalServerError)
		return
	}

	var req models.ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.Orchestrator.ProcessChatRequest(r.Context(), &req, false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logger.Printf("Error encoding chat response: %v", err)
	}
}

// HandleGenerateJourney is a placeholder for journey generation.
func (h *APIHandler) HandleGenerateJourney(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// HandlePreviewJourney is a placeholder for journey preview.
func (h *APIHandler) HandlePreviewJourney(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// HandleUpdateDelays is a placeholder for updating delays.
func (h *APIHandler) HandleUpdateDelays(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// HandleConfirmJourney is a placeholder for journey confirmation.
func (h *APIHandler) HandleConfirmJourney(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// HandleGenerateStep is a placeholder for step generation.
func (h *APIHandler) HandleGenerateStep(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}
