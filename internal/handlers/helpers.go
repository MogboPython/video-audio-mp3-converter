package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func isValidFileType(fileType string) bool {
	validTypes := map[string]bool{
		"mp4": true,
		"mov": true,
		"avi": true,
		"wmv": true,
		"wav": true,
		"m4a": true,
		"aac": true,
		"ogg": true,
		"mp3": true,
	}
	return validTypes[strings.ToLower(fileType)]
}

func downloadFile(url, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download file: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save file: %v", err)
	}

	return nil
}

func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %v", err)
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %v", err)
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	if err != nil {
		return fmt.Errorf("failed to copy file: %v", err)
	}

	return nil
} 