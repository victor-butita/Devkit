package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/victor-butita/devkit/internal/handlers" // Use your module path
	"github.com/victor-butita/devkit/internal/services" // Use your module path
)

func main() {
	// --- Initialization ---
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment")
	}
	geminiAPIKey := os.Getenv("GEMINI_API_KEY")
	if geminiAPIKey == "" {
		log.Fatal("FATAL: GEMINI_API_KEY environment variable is not set")
	}

	// --- Dependency Injection ---
	mockStore := services.NewMockStore()
	geminiService := services.NewGeminiService(geminiAPIKey)
	apiHandlers := handlers.NewAPIHandlers(mockStore, geminiService)

	// --- Routing with Gorilla Mux ---
	r := mux.NewRouter()

	// API routes
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/mock/create", apiHandlers.HandleCreateMock).Methods("POST")
	api.HandleFunc("/regex/generate", apiHandlers.HandleGenerateRegex).Methods("POST")
	api.HandleFunc("/config/convert", apiHandlers.HandleConvertConfig).Methods("POST")
	api.HandleFunc("/sql/generate", apiHandlers.HandleGenerateSQL).Methods("POST")
	api.HandleFunc("/json/format", apiHandlers.HandleFormatJSON).Methods("POST")
	
	// Mock retrieval route - This is a special case outside the /api prefix
	r.HandleFunc("/mock/{id}", apiHandlers.HandleGetMock).Methods("GET")

	// Static file serving for the frontend
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/")))
	
	// --- Server Setup ---
	port := "8080"
	srv := &http.Server{
		Handler:      r,
		Addr:         ":" + port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	fmt.Printf("🚀 DevKit server starting on http://localhost:%s\n", port)
	log.Fatal(srv.ListenAndServe())
}