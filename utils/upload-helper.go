package utils

import (
	"fmt"
	"os"
)

// getUploadBasePath returns the correct base path for uploads
func GetUploadBasePath() string {
	// Check if we're running in production (Docker)
	if _, err := os.Stat("/app"); err == nil {
		return "/app/uploads"
	}
	// Default for local development
	return "uploads"
}

// ensureUploadDir creates upload directory if it doesn't exist
func EnsureUploadDir(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create upload directory: %v", err)
		}
	}
	return nil
}