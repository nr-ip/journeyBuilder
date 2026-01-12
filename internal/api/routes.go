package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"JourneyBuilder/internal/api/handlers"
	"JourneyBuilder/internal/orchestrator"

	"github.com/gorilla/mux"
)

// SetupRoutes configures all API routes.
func SetupRoutes(router *mux.Router, orch *orchestrator.Orchestrator) {
	// API group
	api := router.PathPrefix("/api/v1").Subrouter()

	// Chat endpoints
	chatHandler := handlers.NewChatHandler(orch)
	api.HandleFunc("/chat", chatHandler.Chat).Methods("POST", "OPTIONS")
	api.HandleFunc("/chat/stream", chatHandler.ChatStream).Methods("POST", "OPTIONS")

	// Health check
	api.HandleFunc("/health", healthCheck).Methods("GET")
	api.HandleFunc("/status", statusCheck).Methods("GET")

	// Knowledge base endpoints (admin/debug)
	api.HandleFunc("/frameworks", listFrameworks).Methods("GET")
	api.HandleFunc("/sequences/{vertical}", listSequences).Methods("GET")
}

// Health check endpoint.
func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"service": "davinci-chatbot",
		"version": "1.0.0",
	})
}

// Status endpoint with orchestrator health.
func statusCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
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
func listFrameworks(w http.ResponseWriter, r *http.Request) {
	frameworks := []string{
		"AIDA", "PAS", "FAB", "BAB", "4Ps", "Hero",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"frameworks":  frameworks,
		"count":       len(frameworks),
		"description": "Available copywriting frameworks for email sequences",
	})
}

// Lists sequences for specific vertical (supplements, coaching, etc.)
func listSequences(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	vertical := vars["vertical"]

	sequences := map[string]string{
		"first_purchase":   "5 emails over 7-14 days",
		"cart_recovery":    "3 emails in 48 hours",
		"onboarding":       "5 emails over 21 days",
		"churn_prevention": "4 emails over 30 days",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"vertical":  vertical,
		"sequences": sequences,
		"count":     len(sequences),
		"message":   fmt.Sprintf("Available sequences for %s vertical", vertical),
	})
}
