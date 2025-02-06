package utils

import (
	"fmt"
	"log"
	"os/exec"
	"time"
)

func GetTimeInMillisec() string {
	return time.Now().UTC().Format("20060102_150405")
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
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("FFmpeg error: %s", string(output))
		return fmt.Errorf("ffmpeg failed: %v", err)
	}
	return nil
}
