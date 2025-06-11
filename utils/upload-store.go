// // utils/upload.go
// package utils

// import (
// 	"fmt"
// 	"mime/multipart"
// 	"os"
// 	"path/filepath"
// 	"strings"

// 	"github.com/gin-gonic/gin"
// 	"github.com/google/uuid"
// )

// // FileValidator memeriksa tipe dan ukuran file
// func FileValidator(file *multipart.FileHeader, maxSize int64) error {
// 	if file.Size > maxSize {
// 		return fmt.Errorf("ukuran file melebihi batas %d byte", maxSize)
// 	}

// 	ext := filepath.Ext(file.Filename)
// 	validExts := map[string]bool{
// 		".jpg":  true,
// 		".jpeg": true,
// 		".png":  true,
// 		".gif":  true,
// 		".webp": true,
// 	}

// 	if !validExts[strings.ToLower(ext)] {
// 		return fmt.Errorf("tipe file tidak valid, hanya JPG, JPEG, PNG, GIF, dan WEBP yang diizinkan")
// 	}

// 	return nil
// }

// // UploadFile menyimpan file ke direktori yang ditentukan
// func UploadFile(c *gin.Context, file *multipart.FileHeader, directory string) (string, error) {
// 	// Pastikan direktori ada
// 	if _, err := os.Stat(directory); os.IsNotExist(err) {
// 		if err := os.MkdirAll(directory, 0755); err != nil {
// 			return "", fmt.Errorf("gagal membuat direktori: %v", err)
// 		}
// 	}

// 	// Buat nama file unik
// 	ext := filepath.Ext(file.Filename)
// 	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
// 	filepath := filepath.Join(directory, filename)

// 	// Simpan file
// 	if err := c.SaveUploadedFile(file, filepath); err != nil {
// 		return "", fmt.Errorf("gagal menyimpan file: %v", err)
// 	}

// 	// Kembalikan path relatif untuk disimpan di database
// 	return fmt.Sprintf("/%s/%s", directory, filename), nil
// }

// // DeleteFileIfExists menghapus file jika ada
// func DeleteFileIfExists(filePath string) {
// 	trimmedPath := strings.TrimPrefix(filePath, "/")
// 	if _, err := os.Stat(trimmedPath); err == nil {
// 		os.Remove(trimmedPath)
// 	}
// }

package utils

import (
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// FileValidator memeriksa tipe dan ukuran file
func FileValidator(file *multipart.FileHeader, maxSize int64) error {
	if file.Size > maxSize {
		return fmt.Errorf("ukuran file melebihi batas %d byte", maxSize)
	}

	ext := filepath.Ext(file.Filename)
	validExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
	}

	if !validExts[strings.ToLower(ext)] {
		return fmt.Errorf("tipe file tidak valid, hanya JPG, JPEG, PNG, GIF, dan WEBP yang diizinkan")
	}

	return nil
}

// UploadFile menyimpan file ke direktori yang ditentukan
func UploadFile(c *gin.Context, file *multipart.FileHeader, directory string) (string, error) {
	// Tentukan base path berdasarkan environment
	var basePath string
	if _, err := os.Stat("/app"); err == nil {
		// Production (Docker)
		basePath = "/app/uploads"
	} else {
		// Local development
		basePath = "uploads"
	}
	
	// Create full directory path
	var fullDir string
	if strings.HasPrefix(directory, "uploads/") {
		// Remove "uploads/" prefix and use basePath
		relativePath := strings.TrimPrefix(directory, "uploads/")
		fullDir = filepath.Join(basePath, relativePath)
	} else {
		fullDir = filepath.Join(basePath, directory)
	}
	
	// Pastikan direktori ada
	if _, err := os.Stat(fullDir); os.IsNotExist(err) {
		if err := os.MkdirAll(fullDir, 0755); err != nil {
			return "", fmt.Errorf("gagal membuat direktori: %v", err)
		}
	}

	// Buat nama file unik
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	fullPath := filepath.Join(fullDir, filename)

	// Simpan file
	if err := c.SaveUploadedFile(file, fullPath); err != nil {
		return "", fmt.Errorf("gagal menyimpan file: %v", err)
	}

	// Kembalikan path relatif untuk URL (always use forward slashes)
	var urlPath string
	if strings.HasPrefix(directory, "uploads/") {
		urlPath = fmt.Sprintf("/%s/%s", directory, filename)
	} else {
		urlPath = fmt.Sprintf("/uploads/%s/%s", directory, filename)
	}
	
	return urlPath, nil
}

// DeleteFileIfExists menghapus file jika ada
func DeleteFileIfExists(filePath string) {
	if filePath == "" {
		return
	}
	
	// Tentukan base path berdasarkan environment
	var basePath string
	if _, err := os.Stat("/app"); err == nil {
		// Production (Docker)
		basePath = "/app/uploads"
	} else {
		// Local development
		basePath = "uploads"
	}
	
	// Convert URL path to actual file path
	var actualPath string
	if strings.HasPrefix(filePath, "/uploads/") {
		// Remove "/uploads/" prefix and use basePath
		relativePath := strings.TrimPrefix(filePath, "/uploads/")
		actualPath = filepath.Join(basePath, relativePath)
	} else {
		// Handle legacy paths
		trimmedPath := strings.TrimPrefix(filePath, "/")
		actualPath = trimmedPath
	}
	
	if _, err := os.Stat(actualPath); err == nil {
		os.Remove(actualPath)
	}
}