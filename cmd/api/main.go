package main

import (
	"JourneyBuilder/internal/api/handlers"
	"JourneyBuilder/internal/knowledge"
	"JourneyBuilder/internal/logger"
	"JourneyBuilder/internal/orchestrator"
	"JourneyBuilder/internal/services"
	"JourneyBuilder/internal/validation"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

func main() {
	// Load .env file if it exists (doesn't override existing env vars)
	if err := godotenv.Load(); err != nil {
		log.Println("Info: .env file not found, using system environment variables")
	} else {
		log.Println("✓ Loaded .env file")
	}

	// Initialize logger (file + console logging)
	if err := logger.InitLogger(); err != nil {
		log.Printf("Warning: Failed to initialize logger: %v. Continuing with console logging only.", err)
	} else {
		logFile := os.Getenv("LOG_FILE")
		if logFile != "" {
			log.Printf("✓ Logging to file: %s", logFile)
		} else {
			log.Println("ℹ️  Logging to console only (set LOG_FILE env var to enable file logging)")
		}
	}
	defer logger.Close()

	// Verify required environment variables
	//checkRequiredEnvVars()  // check TODO in function definition

	port := os.Getenv("FRONTEND_PORT")
	if port == "" {
		port = "8080"
	}

	// Using custom logger from internal/logger.
	logger.Println("Initializing AI services...")
	geminiService, err := services.NewGeminiService()
	if err != nil {
		logger.Printf("Failed to initialize Gemini service: %v", err)
		return
	}
	defer geminiService.Close()

	// Initialize knowledge base
	frameworksPath := filepath.Join("data", "knowledge", "frameworks.json")
	sequencesPath := filepath.Join("data", "knowledge", "sequence.json")
	verticalsPath := filepath.Join("data", "knowledge", "verticals.json")
	kb, err := knowledge.NewKnowledgeBase(frameworksPath, sequencesPath, verticalsPath)
	if err != nil {
		logger.Printf("Failed to initialize knowledge base: %v", err)
		return
	}

	// Initialize validation
	inputValidator := validation.NewInputValidator()
	outputValidator := validation.NewOutputValidator()

	// Initialize orchestrator
	orch := orchestrator.NewOrchestrator(geminiService, kb, inputValidator, outputValidator)

	// Create a new APIHandler instance and inject the orchestrator
	apiHandler := handlers.NewAPIHandler(orch)

	router := mux.NewRouter()

	// Health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"status":"ok","service":"journey-builder"}`)); err != nil {
			logger.Printf("Error writing health check response: %v", err)
		}
	}).Methods("GET")

	// API routes
	router.HandleFunc("/api/chat", apiHandler.HandleChat).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/generate-journey", apiHandler.HandleGenerateJourney).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/preview-journey", apiHandler.HandlePreviewJourney).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/update-delays", apiHandler.HandleUpdateDelays).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/confirm-journey", apiHandler.HandleConfirmJourney).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/generate-step", apiHandler.HandleGenerateStep).Methods("POST", "OPTIONS")

	// Serve static files from public directory (must be last to catch all other routes)
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./public")))

	// Setup CORS with dynamic origins from environment variable
	corsOriginsStr := os.Getenv("CORS_ALLOWED_ORIGINS")
	var allowedOrigins []string
	if corsOriginsStr == "" {
		logger.Println("Warning: CORS_ALLOWED_ORIGINS environment variable not set. Cross-origin requests will be blocked.")
		allowedOrigins = []string{
			"https://staging.maropost.com",
			"https://uat.maropost.com",
			"https://app.maropost.com", //Defaulting to application links.
		}
	} else {
		allowedOrigins = strings.Split(corsOriginsStr, ",")
		logger.Printf("✓ CORS origins loaded: %v", allowedOrigins)
	}

	c := cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"}, // More specific than "*"
		AllowCredentials: true,
	})
	handler := c.Handler(router)

	server := &http.Server{Addr: ":" + port, Handler: handler}

	// Server run context
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	// Listen for syscall signals for process interruption.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig

		// Shutdown signal with grace period of 30 seconds.
		// Use Background() to ensure the shutdown process has its own dedicated window.
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Trigger graceful shutdown
		logger.Println("\n🛑 Shutting down gracefully...")
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Printf("⚠️ Server shutdown error: %v", err)
		} else {
			logger.Println("✓ HTTP server stopped successfully")
		}
		serverStopCtx()
	}()

	// Run the server
	logger.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	logger.Printf("🚀 Server starting on port %s", port)
	logger.Printf("📱 Open http://localhost:%s in your browser", port)
	logger.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		logger.Printf("ListenAndServe error: %v", err)
		serverStopCtx()
	}

	// Wait for server context to be stopped
	<-serverCtx.Done()
	logger.Println("✓ Process cleanup complete. Exiting.")
}
