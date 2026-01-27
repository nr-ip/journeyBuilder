package main

import (
	"JourneyBuilder/internal/api/handlers"
	"JourneyBuilder/internal/knowledge"
	"JourneyBuilder/internal/logger"
	"JourneyBuilder/internal/orchestrator"
	"JourneyBuilder/internal/services"
	"JourneyBuilder/internal/validation"

	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

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

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Using custom logger from internal/logger.
	logger.Println("Initializing AI services...")
	geminiService, err := services.NewGeminiService()
	if err != nil {
		log.Fatalf("Failed to initialize Gemini service: %v", err)
	}
	defer geminiService.Close()

	// Initialize knowledge base
	frameworksPath := filepath.Join("data", "knowledge", "frameworks.json")
	sequencesPath := filepath.Join("data", "knowledge", "sequence.json")
	verticalsPath := filepath.Join("data", "knowledge", "verticals.json")
	kb, err := knowledge.NewKnowledgeBase(frameworksPath, sequencesPath, verticalsPath)
	if err != nil {
		logger.Fatalf("Failed to initialize knowledge base: %v", err)
	}

	// Initialize validation
	inputValidator := validation.NewInputValidator()
	outputValidator := validation.NewOutputValidator()

	// Initialize orchestrator
	orch := orchestrator.NewOrchestrator(geminiService, kb, inputValidator, outputValidator)
	handlers.SetOrchestrator(orch)
	setupGracefulShutdown(geminiService)

	router := mux.NewRouter()

	// Health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"journey-builder"}`))
	}).Methods("GET")

	// API routes
	router.HandleFunc("/api/chat", handlers.HandleChat).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/generate-journey", handlers.HandleGenerateJourney).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/preview-journey", handlers.HandlePreviewJourney).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/update-delays", handlers.HandleUpdateDelays).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/confirm-journey", handlers.HandleConfirmJourney).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/generate-step", handlers.HandleGenerateStep).Methods("POST", "OPTIONS")

	// Serve static files from public directory (must be last to catch all other routes)
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./public")))

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})
	handler := c.Handler(router)
	logger.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	logger.Printf("🚀 Server starting on port %s", port)
	logger.Printf("📱 Open http://localhost:%s in your browser", port)
	logger.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal(err)
	}
}

// TODO: Will optimize this later. Probably moved to gemini.go
// checkRequiredEnvVars verifies that required environment variables are set
// func checkRequiredEnvVars() {
// 	required := []string{"GCP_PROJECT_ID", "GCP_REGION", "GEMINI_MODEL"}
// 	missing := []string{}

// 	for _, key := range required {
// 		value := os.Getenv(key)
// 		if value == "" {
// 			missing = append(missing, key)
// 		} else {
// 			// Mask the value for security (show first 4 chars)
// 			masked := value
// 			if len(value) > 4 {
// 				masked = value[:4] + "..."
// 			}
// 			log.Printf("✓ Found %s: %s", key, masked)
// 		}
// 	}

// 	if len(missing) > 0 {
// 		log.Printf("⚠️  Missing required environment variables: %v", missing)
// 		log.Println("   Set them in your .env file or system environment:")
// 		for _, key := range missing {
// 			log.Printf("   export %s=your_value_here", key)
// 		}
// 	}

// 	// Check optional variables
// 	optional := map[string]string{
// 		"GEMINI_MODEL":                   "gemini-2.5-flash (default)",
// 		"GOOGLE_APPLICATION_CREDENTIALS": "not set (optional, for Vertex AI)",
// 	}

// 	for key, defaultValue := range optional {
// 		value := os.Getenv(key)
// 		if value == "" {
// 			log.Printf("ℹ️  %s: %s", key, defaultValue)
// 		} else {
// 			log.Printf("✓ Found %s: %s", key, value)
// 		}
// 	}
// }

func setupGracefulShutdown(geminiService *services.GeminiService) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		logger.Println("\n🛑 Shutting down gracefully...")
		if err := geminiService.Close(); err != nil {
			logger.Printf("Error closing Gemini service: %v", err)
		}
		log.Println("✓ Cleanup complete")
		os.Exit(0)
	}()
}
