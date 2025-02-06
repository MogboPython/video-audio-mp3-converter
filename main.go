package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	googleStorage "cloud.google.com/go/storage"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/storage"
	"github.com/MogboPython/video-audio-mp3-converter/utils"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"google.golang.org/api/option"
)

type ConversionResponse struct {
	URL     string `json:"url"`
	Message string `json:"message"`
}

type StreamHandler struct {
	firebaseStorage *storage.Client
	bucket          string
}

func NewStreamHandler(firebaseStorage *storage.Client, bucket string) *StreamHandler {
	return &StreamHandler{
		firebaseStorage: firebaseStorage,
		bucket:          bucket,
	}
}

func (h *StreamHandler) HandleUpload(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("Content-Type", "application/json")

	// Get user and meeting IDs from headers
	userId := r.Header.Get("X-User-ID")
	meetingId := r.Header.Get("X-Meeting-ID")
	tempUrl := r.Header.Get("X-Temp-URL")

	// Validate required headers
	if userId == "" || meetingId == "" {
		http.Error(w, "Missing required headers: X-User-ID and X-Meeting-ID", http.StatusBadRequest)
		return
	}

	// Generate timestamp for filename
	// timestamp := time.Now().UTC().Format("20060102_150405")
	timestamp := utils.GetTimeInMillisec()

	// Create the Firebase Storage path
	storagePath := fmt.Sprintf("users/%s/meetings/%s/audio_%s.mp3", userId, meetingId, timestamp)

	log.Printf("Processing upload for user %s, meeting %s", userId, meetingId)

	// Create temporary file
	tempID := uuid.New().String()
	inputPath := filepath.Join(os.TempDir(), tempID)
	outputPath := inputPath + ".mp3"

	tempFile, err := os.Create(inputPath)
	if err != nil {
		log.Printf("Failed to create temp file: %v", err)
		http.Error(w, "Failed to create temp file", http.StatusInternalServerError)
		return
	}
	defer os.Remove(inputPath)
	defer tempFile.Close()

	// Read the stream in chunks
	buffer := make([]byte, 32*1024) // 32KB chunks
	for {
		n, err := r.Body.Read(buffer)
		if n > 0 {
			if _, writeErr := tempFile.Write(buffer[:n]); writeErr != nil {
				log.Printf("Failed to write chunk: %v", writeErr)
				http.Error(w, "Failed to write chunk", http.StatusInternalServerError)
				return
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Failed to read chunk: %v", err)
			http.Error(w, "Failed to read chunk", http.StatusBadRequest)
			return
		}
	}

	// Ensure all data is written
	if err := tempFile.Sync(); err != nil {
		log.Printf("Failed to sync file: %v", err)
		http.Error(w, "Failed to sync file", http.StatusInternalServerError)
		return
	}

	// Convert to MP3 using ffmpeg
	if err := utils.ConvertToMp3(inputPath, outputPath); err != nil {
		log.Printf("Conversion failed: %v", err)
		http.Error(w, "Conversion failed", http.StatusInternalServerError)
		return
	}
	defer os.Remove(outputPath)

	// Upload to Firebase Storage
	// ctx := context.Background()
	outputFile, err := os.Open(outputPath)
	if err != nil {
		log.Printf("Failed to read converted file: %v", err)
		http.Error(w, "Failed to read converted file", http.StatusInternalServerError)
		return
	}
	defer outputFile.Close()

	// Create storage object with metadata
	bucketHandle, err := h.firebaseStorage.Bucket(h.bucket)
	if err != nil {
		// Handle the error, e.g., log it and return or panic
		log.Fatalf("Failed to get bucket handle: %v", err)
	}

	obj := bucketHandle.Object(storagePath)
	writer := obj.NewWriter(ctx)
	writer.ContentType = "audio/mp3" // audio/mpeg
	// FIXME: date part of meta data
	writer.Metadata = map[string]string{
		"userId":    userId,
		"meetingId": meetingId,
		"timestamp": timestamp,
		"tempUrl":   tempUrl,
	}

	// Copy file to storage
	if _, err := io.Copy(writer, outputFile); err != nil {
		log.Printf("Failed to upload to storage: %v", err)
		http.Error(w, "Failed to upload to storage", http.StatusInternalServerError)
		return
	}

	// Close the writer
	if err := writer.Close(); err != nil {
		log.Printf("Failed to finalize upload: %v", err)
		http.Error(w, "Failed to finalize upload", http.StatusInternalServerError)
		return
	}

	// FIXME: here
	// Get the public URL
	opts := &googleStorage.SignedURLOptions{
		GoogleAccessID: os.Getenv("FIREBASE_CLIENT_EMAIL"),
		PrivateKey:     []byte(os.Getenv("FIREBASE_PRIVATE_KEY")),
		Method:         "GET",
		Expires:        time.Now().Add(7 * 24 * time.Hour),
	}

	url, err := bucketHandle.SignedURL(storagePath, opts)

	if err != nil {
		log.Printf("Failed to generate URL: %v", err)
		http.Error(w, "Failed to generate URL", http.StatusInternalServerError)
		return
	}
	// url := storagePath

	// Send response
	json.NewEncoder(w).Encode(ConversionResponse{
		URL:     url,
		Message: "Conversion successful",
	})
}

func main() {
	// Initialize Firebase
	ctx := context.Background()
	config := &firebase.Config{
		StorageBucket: os.Getenv("FIREBASE_STORAGE_BUCKET"),
	}

	opt := option.WithCredentialsFile(os.Getenv("CREDENTIALS"))
	app, err := firebase.NewApp(ctx, config, opt)
	if err != nil {
		log.Fatalf("Failed to initialize Firebase: %v", err)
	}

	storage, err := app.Storage(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// Initialize router
	r := mux.NewRouter()

	// Initialize handler
	handler := NewStreamHandler(storage, os.Getenv("FIREBASE_STORAGE_BUCKET"))
	// Set up routes
	r.HandleFunc("/convert", func(w http.ResponseWriter, r *http.Request) {
		handler.HandleUpload(ctx, w, r)
	}).Methods("POST")

	// Set up CORS
	c := cors.New(cors.Options{
		AllowedOrigins: []string{os.Getenv("ALLOWED_ORIGIN")}, // Your Vercel app URL
		AllowedMethods: []string{"POST", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type"},
		MaxAge:         86400, // 24 hours
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, c.Handler(r)))
}
