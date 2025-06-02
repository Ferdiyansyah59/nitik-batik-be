// utils/upload.go
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
func FileValidatorProduct(file *multipart.FileHeader, maxSize int64) error {
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
func UploadFileproduct(c *gin.Context, file *multipart.FileHeader, directory string) (string, error) {
	// Pastikan direktori ada
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		if err := os.MkdirAll(directory, 0755); err != nil {
			return "", fmt.Errorf("gagal membuat direktori: %v", err)
		}
	}

	// Buat nama file unik
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	filepath := filepath.Join(directory, filename)

	// Simpan file
	if err := c.SaveUploadedFile(file, filepath); err != nil {
		return "", fmt.Errorf("gagal menyimpan file: %v", err)
	}

	// Kembalikan path relatif untuk disimpan di database
	return fmt.Sprintf("/%s/%s", directory, filename), nil
}

// DeleteFileIfExists menghapus file jika ada
func DeleteFileIfExistsProduct(filePath string) {
	if filePath == "" {
		return
	}
	
	trimmedPath := strings.TrimPrefix(filePath, "/")
	if _, err := os.Stat(trimmedPath); err == nil {
		os.Remove(trimmedPath)
	}
}