// package utils

// import (
// 	"batik/entity"
// 	"batik/helper"
// 	"fmt"
// 	"net/http"
// 	"os"
// 	"path/filepath"
// 	"time"

// 	"github.com/gin-gonic/gin"
// 	"github.com/google/uuid"
// )

// // UploadImageResponse represents the response from the image upload endpoint

// // UploadImage handles image uploads for articles
// func UploadImage(c *gin.Context) {
// 	// Get file from request
// 	file, err := c.FormFile("image")
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "No image provided"})
// 		return
// 	}

// 	// Check file size (limit to 5MB)
// 	if file.Size > 5*1024*1024 {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Image size exceeds 5MB limit"})
// 		return
// 	}

// 	// Check file extension
// 	ext := filepath.Ext(file.Filename)
// 	validExts := map[string]bool{
// 		".jpg":  true,
// 		".jpeg": true,
// 		".png":  true,
// 		".gif":  true,
// 		".webp": true,
// 	}

// 	if !validExts[ext] {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Only JPG, JPEG, PNG, GIF, and WEBP are allowed"})
// 		return
// 	}

// 	// Ensure uploads directory exists
// 	uploadDir := "uploads/images"
// 	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
// 		if err := os.MkdirAll(uploadDir, 0755); err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
// 			return
// 		}
// 	}

// 	// Generate a unique filename
// 	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
// 	filepath := filepath.Join(uploadDir, filename)

// 	// Save the file
// 	if err := c.SaveUploadedFile(file, filepath); err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image"})
// 		return
// 	}

// 	// Return the file path
// 	imageURL := fmt.Sprintf("/uploads/images/%s", filename)
// 	response := entity.UploadImageResponse{
// 		ImageURL:   imageURL,
// 		Filename:   filename,
// 		Size:       file.Size,
// 		UploadedAt: time.Now(),
// 	}

// 	c.JSON(http.StatusOK, helper.BuildResponse(true, "Image uploaded successfully", response))
// }

package utils

import (
	"batik/entity"
	"batik/helper"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UploadImage handles image uploads for articles
func UploadImage(c *gin.Context) {
	// Get file from request
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No image provided"})
		return
	}

	// Check file size (limit to 5MB)
	if file.Size > 5*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Image size exceeds 5MB limit"})
		return
	}

	// Check file extension
	ext := filepath.Ext(file.Filename)
	validExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
	}

	if !validExts[strings.ToLower(ext)] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Only JPG, JPEG, PNG, GIF, and WEBP are allowed"})
		return
	}

	// Determine base path based on environment
	var basePath string
	if _, err := os.Stat("/app"); err == nil {
		// Production (Docker)
		basePath = "/app/uploads"
	} else {
		// Local development
		basePath = "uploads"
	}
	
	// Create full directory path
	uploadDir := filepath.Join(basePath, "images")
	
	// Ensure uploads directory exists
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
			return
		}
	}

	// Generate a unique filename
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	fullPath := filepath.Join(uploadDir, filename)

	// Save the file
	if err := c.SaveUploadedFile(file, fullPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image"})
		return
	}

	// Return the file path (always use forward slashes for URL)
	imageURL := fmt.Sprintf("/uploads/images/%s", filename)
	response := entity.UploadImageResponse{
		ImageURL:   imageURL,
		Filename:   filename,
		Size:       file.Size,
		UploadedAt: time.Now(),
	}

	c.JSON(http.StatusOK, helper.BuildResponse(true, "Image uploaded successfully", response))
}


