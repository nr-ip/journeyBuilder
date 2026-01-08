package handlers

import (
	"JourneyBuilder/internal/services"
	"log"
)

var geminiService *services.GeminiService

func InitializeGeminiService() error {
	var err error
	geminiService, err = services.NewGeminiService()
	if err != nil {
		return err
	}
	log.Println("✓ Gemini AI service initialized")
	return nil
}

func CloseGeminiService() {
	if geminiService != nil {
		geminiService.Close()
	}
}
