package services

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	googleStorage "cloud.google.com/go/storage"
	"firebase.google.com/go/v4/storage"
	"github.com/MogboPython/video-audio-mp3-converter/pkg/utils"
)

type FirebaseStorageService struct {
	firebaseStorage *storage.Client
	bucket          string
}

func NewFirebaseStorageService(firebaseStorage *storage.Client, bucket string) *FirebaseStorageService {
	return &FirebaseStorageService{
		firebaseStorage: firebaseStorage,
		bucket:          bucket,
	}
}

func (h *FirebaseStorageService) UploadFile(ctx context.Context, path, userId, meetingId, tempUrl string, w http.ResponseWriter) (string, error) {
	outputFile, err := os.Open(path)
	if err != nil {
		log.Printf("Failed to read converted file: %v", err)
		return "", fmt.Errorf("Failed to read converted file")
	}
	defer outputFile.Close()

	// Create the Firebase Storage path
	storagePath := fmt.Sprintf("users/%s/meetings/%s/audio_%s.mp3", userId, meetingId, utils.GetTimeInMillisec())

	bucketHandle, err := h.firebaseStorage.Bucket(h.bucket)
	if err != nil {
		log.Printf("Failed to get bucket handle: %v", err)
		return "", fmt.Errorf("Failed to get bucket handle")
	}

	obj := bucketHandle.Object(storagePath)
	writer := obj.NewWriter(ctx)
	writer.ContentType = "audio/mp3"
	writer.Metadata = map[string]string{
		"userId":    userId,
		"meetingId": meetingId,
		"timestamp": utils.GetCurrentTimeISO(),
		"tempUrl":   tempUrl,
	}

	// Copy file to storage
	if _, err := io.Copy(writer, outputFile); err != nil {
		log.Printf("Failed to upload to storage: %v", err)
		return "", fmt.Errorf("Failed to upload to storage")
	}

	// Close the writer
	if err := writer.Close(); err != nil {
		log.Printf("Failed to finalize upload: %v", err)
		return "", fmt.Errorf("Failed to finalize upload")

	}

	return storagePath, nil
}

// Get the public URL
func (h *FirebaseStorageService) GenerateSignedURL(path string) (string, error) {
	bucketHandle, err := h.firebaseStorage.Bucket(h.bucket)
	if err != nil {
		log.Printf("Failed to get bucket handle: %v", err)
		return "", fmt.Errorf("Failed to get bucket handle")
	}

	// FIXME: use config
	opts := &googleStorage.SignedURLOptions{
		GoogleAccessID: os.Getenv("FIREBASE_CLIENT_EMAIL"),
		PrivateKey:     []byte(os.Getenv("FIREBASE_PRIVATE_KEY")),
		Method:         "GET",
		Expires:        time.Now().Add(7 * 24 * time.Hour),
	}

	url, err := bucketHandle.SignedURL(path, opts)

	if err != nil {
		log.Printf("Failed to generate URL: %v", err)
		return "", fmt.Errorf("Failed to generate URL")
	}

	return url, nil
}

// DeleteFile deletes a file from Firebase Storage
func (h *FirebaseStorageService) DeleteFile(ctx context.Context, fileURL string) error {
	// Extract path from Firebase Storage URL
	path, err := extractPathFromURL(fileURL)
	if err != nil {
		return fmt.Errorf("failed to parse URL: %w", err)
	}

	bucketHandle, err := h.firebaseStorage.Bucket(h.bucket)
	if err != nil {
		return fmt.Errorf("failed to get bucket handle: %w", err)
	}

	obj := bucketHandle.Object(path)
	if err := obj.Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// Helper function to extract path from Firebase Storage URL
func extractPathFromURL(fileURL string) (string, error) {
	// Parse the URL
	u, err := url.Parse(fileURL)
	if err != nil {
		return "", err
	}

	// The path is in the 'o' parameter
	path, err := url.QueryUnescape(u.Path)
	if err != nil {
		return "", err
	}

	// Remove the '/o/' prefix if it exists
	path = strings.TrimPrefix(path, "/v0/b/")
	path = strings.TrimPrefix(path, fmt.Sprintf("%s/o/", u.Host))
	
	return path, nil
}
