package utils

import (
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
)

// DetectContentTypeFromFile reads the first 512 bytes of the file and detects its MIME type.
func DetectContentTypeFromFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Read the first 512 bytes to detect the content type
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", err
	}

	// Detect the content type
	contentType := http.DetectContentType(buffer)

	// If the detected content type is "application/octet-stream",
	// fall back to the file extension
	if contentType == "application/octet-stream" {
		contentType = mime.TypeByExtension(filepath.Ext(filePath))
		if contentType == "" {
			contentType = "application/octet-stream" // Default MIME type
		}
	}

	return contentType, nil
}
