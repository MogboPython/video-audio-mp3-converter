package handlers

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
)

func (h *StreamHandler) updateMeetingDoc(ctx context.Context, meetingId, audioPath string) error {
	meetingRef := h.firestoreClient.Collection("meetings").Doc(meetingId)
	
	// First get the signed URL
	signedURL, err := h.storageService.GenerateSignedURL(audioPath)
	if err != nil {
		return fmt.Errorf("failed to generate signed URL: %v", err)
	}

	updates := []firestore.Update{
		{
			Path:  "audioUrl",
			Value: signedURL,
		},
		{
			Path:  "audioPath",
			Value: audioPath,
		},
		{
			Path:  "audioUrlLastRefresh",
			Value: firestore.ServerTimestamp,
		},
		{
			Path:  "audio_processed_at",
			Value: firestore.ServerTimestamp,
		},
	}

	_, err = meetingRef.Update(ctx, updates)
	if err != nil {
		return fmt.Errorf("failed to update meeting document: %v", err)
	}

	return nil
} 