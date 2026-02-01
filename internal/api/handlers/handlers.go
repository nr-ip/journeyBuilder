package handlers

import (
	"encoding/json"
	"net/http"

	"JourneyBuilder/internal/models"
	"JourneyBuilder/internal/orchestrator"
)

var globalOrchestrator *orchestrator.Orchestrator

// SetOrchestrator sets the global orchestrator instance
func SetOrchestrator(orch *orchestrator.Orchestrator) {
	globalOrchestrator = orch
}

// HandleChat handles chat requests using the global orchestrator
func HandleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if globalOrchestrator == nil {
		http.Error(w, "Orchestrator not initialized", http.StatusInternalServerError)
		return
	}

	var req models.ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := globalOrchestrator.ProcessChatRequest(r.Context(), &req, false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// HandleGenerateJourney is a placeholder for journey generation
func HandleGenerateJourney(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// HandlePreviewJourney is a placeholder for journey preview
func HandlePreviewJourney(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// HandleUpdateDelays is a placeholder for updating delays
func HandleUpdateDelays(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// HandleConfirmJourney is a placeholder for journey confirmation
func HandleConfirmJourney(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// HandleGenerateStep is a placeholder for step generation
func HandleGenerateStep(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}
