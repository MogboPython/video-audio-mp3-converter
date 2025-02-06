package utils

import (
	"fmt"
	"os/exec"
	"time"
)

func GetTimeInMillisec() string {
	return time.Now().UTC().Format("20060102_150405")
}

// GetCurrentTimeISO returns the current time in ISO 8601 format.
func GetCurrentTimeISO() string {
	currentTime := time.Now()
	return currentTime.Format(time.RFC3339) // ISO 8601 format
}

func ConvertToMp3(inputPath, outputPath string) error {
	cmd := exec.Command("ffmpeg",
		"-i", inputPath,
		"-f", "mp3", // Output format
		"-vn", // Disable video
		"-acodec", "libmp3lame",
		"-ab", "128k", // Bitrate
		"-ac", "2", // Audio channels
		"-ar", "44100", // Sample rate
		"-y", // Overwrite output
		outputPath,
	)

	// Capture ffmpeg output for debugging
	_, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg error - %v", err)
	}
	return nil
}
