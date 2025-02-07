package main

import (
	"context"
	"log"
	"net/http"
	"os"

	firebase "firebase.google.com/go/v4"
	"cloud.google.com/go/firestore"
	"github.com/MogboPython/video-audio-mp3-converter/internal/config"
	"github.com/MogboPython/video-audio-mp3-converter/internal/handlers"
	"github.com/MogboPython/video-audio-mp3-converter/internal/router"
	"github.com/MogboPython/video-audio-mp3-converter/internal/services"
	"google.golang.org/api/option"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize Firestore client
	ctx := context.Background()
	
	// Get credentials path
	credentialsPath := os.Getenv("CREDENTIALS")
	opt := option.WithCredentialsFile(credentialsPath)
	
	firestoreClient, err := firestore.NewClient(ctx, cfg.ProjectID, opt)
	if err != nil {
		log.Fatalf("Failed to create Firestore client: %v", err)
	}
	defer firestoreClient.Close()

	// Initialize services
	storageService, err := initializeServices(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize services: %v", err)
	}

	// Initialize handlers with both services
	streamHandler := handlers.NewStreamHandler(storageService, firestoreClient)

	// Setup router
	r := router.SetupRouter(streamHandler, cfg.AllowedOrigin)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func initializeServices(cfg *config.Config) (*services.FirebaseStorageService, error) {
	ctx := context.Background()

	config := &firebase.Config{
		ProjectID:     cfg.ProjectID,
		StorageBucket: cfg.StorageBucket,
	}

	app, err := firebase.NewApp(ctx, config, cfg.Opt)
	if err != nil {
		return nil, err
	}

	storage, err := app.Storage(ctx)
	if err != nil {
		return nil, err
	}

	storageService := services.NewFirebaseStorageService(storage, cfg.StorageBucket)
	return storageService, nil
}
