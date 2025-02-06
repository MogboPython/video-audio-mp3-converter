package ports

import (
	"context"
	"net/http"
)

// StorageService defines the interface for storage operations
type StorageService interface {
	UploadFile(ctx context.Context, path, userId, meetingId, tempUrl string, w http.ResponseWriter) (string, error)
	GenerateSignedURL(path string) (string, error)
	DeleteFile(ctx context.Context, path string) error
}
