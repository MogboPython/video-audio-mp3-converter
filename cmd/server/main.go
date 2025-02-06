package main

import (
	"context"
	"log"
	"net/http"

	firebase "firebase.google.com/go/v4"
	"github.com/MogboPython/video-audio-mp3-converter/internal/config"
	"github.com/MogboPython/video-audio-mp3-converter/internal/handlers"
	"github.com/MogboPython/video-audio-mp3-converter/internal/router"
	"github.com/MogboPython/video-audio-mp3-converter/internal/services"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize services
	storageService, err := initializeServices(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize services: %v", err)
	}

	// Initialize handlers
	streamHandler := handlers.NewStreamHandler(storageService)

	// Setup router
	router := router.SetupRouter(streamHandler, cfg)

	// Start server
	log.Printf("Starting server on port %s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, router))
}

func initializeServices(cfg *config.Config) (*services.FirebaseStorageService, error) {
	ctx := context.Background()

	config := &firebase.Config{
		StorageBucket: cfg.StorageBucket,
	}

	app, err := firebase.NewApp(ctx, config, cfg.Opt)
	if err != nil {
		log.Fatalf("Failed to initialize Firebase: %v", err)
	}

	storage, err := app.Storage(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	storageService := services.NewFirebaseStorageService(storage, cfg.StorageBucket)
	return storageService, nil
}
