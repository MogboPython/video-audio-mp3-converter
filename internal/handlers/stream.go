package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"cloud.google.com/go/firestore"
	"github.com/MogboPython/video-audio-mp3-converter/internal/ports"
	"github.com/MogboPython/video-audio-mp3-converter/pkg/utils"
	"github.com/google/uuid"
)

type StreamHandler struct {
	storageService    ports.StorageService
	firestoreClient   *firestore.Client
}

func NewStreamHandler(storageService ports.StorageService, firestoreClient *firestore.Client) *StreamHandler {
	return &StreamHandler{
		storageService:    storageService,
		firestoreClient:   firestoreClient,
	}
}

type Response struct {
	URL     string `json:"audioUrl"`
	Path    string `json:"audioPath"`
	Message string `json:"message"`
}

type RequestPayload struct {
	MeetingId string `json:"meetingId"`
	TempUrl   string `json:"tempUrl"`
	FileType  string `json:"fileType"`
	UserId    string `json:"userId"`
}

func (h *StreamHandler) HandleUpload(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	w.Header().Set("Content-Type", "application/json")

	// Parse request body
	var payload RequestPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if payload.MeetingId == "" || payload.TempUrl == "" || payload.FileType == "" || payload.UserId == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Validate file type
	if !isValidFileType(payload.FileType) {
		http.Error(w, "Unsupported file type", http.StatusBadRequest)
		return
	}

	// Create temporary files with unique names
	tempID := uuid.New().String()
	tempFilePath := filepath.Join(os.TempDir(), tempID+"."+payload.FileType)
	outputPath := filepath.Join(os.TempDir(), tempID+".mp3")

	if err := downloadFile(payload.TempUrl, tempFilePath); err != nil {
		log.Printf("Failed to download file: %v", err)
		http.Error(w, "Failed to download file", http.StatusInternalServerError)
		return
	}
	defer os.Remove(tempFilePath)
	defer os.Remove(outputPath)

	// Convert to MP3 if needed
	if payload.FileType != "mp3" {
		if err := utils.ConvertToMp3(tempFilePath, outputPath); err != nil {
			log.Printf("Conversion failed: %v", err)
			http.Error(w, "Conversion failed", http.StatusInternalServerError)
			return
		}
	} else {
		if err := copyFile(tempFilePath, outputPath); err != nil {
			log.Printf("Failed to copy MP3 file: %v", err)
			http.Error(w, "Failed to process file", http.StatusInternalServerError)
			return
		}
	}

	// Upload to Firebase
	url, err := h.storageService.UploadFile(ctx, outputPath, payload.UserId, payload.MeetingId, payload.TempUrl, w)
	if err != nil {
		http.Error(w, "Failed to upload file", http.StatusInternalServerError)
		return
	}

	// Delete original temp file if it's from Firebase Storage
	if strings.Contains(payload.TempUrl, "firebasestorage.googleapis.com") {
		// Extract path from URL
		path := extractPathFromURL(payload.TempUrl)
		if err := h.storageService.DeleteFile(ctx, path); err != nil {
			log.Printf("Failed to delete temp file: %v", err)
			// Continue execution as this is not critical
		}
	}

	// Update meeting document
	if err := h.updateMeetingDoc(ctx, payload.MeetingId, url); err != nil {
		log.Printf("Failed to update meeting doc: %v", err)
		http.Error(w, "Failed to update meeting", http.StatusInternalServerError)
		return
	}

	signedURL, err := h.storageService.GenerateSignedURL(url)
	if err != nil {
		log.Printf("Failed to generate signed URL: %v", err)
		http.Error(w, "Failed to generate signed URL", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(Response{
		URL:     signedURL,
		Path:    url,
		Message: "Processing completed successfully",
	})
}

// Helper function to extract path
func extractPathFromURL(fileURL string) string {
	// Parse the URL
	u, _ := url.Parse(fileURL)
	
	// Get the 'o' parameter which contains the path
	path := strings.TrimPrefix(u.Path, "/v0/b/xophieai.firebasestorage.app/o/")
	
	// URL decode the path
	path, _ = url.QueryUnescape(path)
	
	return path
}
