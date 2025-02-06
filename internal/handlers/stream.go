package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/MogboPython/video-audio-mp3-converter/internal/ports"
	"github.com/MogboPython/video-audio-mp3-converter/pkg/utils"
	"github.com/google/uuid"
)

type StreamHandler struct {
	storageService ports.StorageService
}

func NewStreamHandler(storageService ports.StorageService) *StreamHandler {
	return &StreamHandler{
		storageService: storageService,
	}
}

type Response struct {
	URL     string `json:"url"`
	Message string `json:"message"`
}

func (h *StreamHandler) HandleUpload(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
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

	if err := utils.ConvertToMp3(inputPath, outputPath); err != nil {
		log.Printf("Conversion failed: %v", err)
		http.Error(w, "Conversion failed", http.StatusInternalServerError)
		return
	}
	defer os.Remove(outputPath)

	url, err := h.storageService.UploadFile(ctx, outputPath, userId, meetingId, tempUrl, w)
	if err != nil {
		http.Error(w, "Failed to upload file", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(Response{
		URL:     url,
		Message: "Conversion successful",
	})

}
