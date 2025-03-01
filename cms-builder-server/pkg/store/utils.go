package store

import (
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// randomizeFileName takes a file name and returns a new file name that is
// randomized using the current timestamp. The file name is also sanitized
// to replace spaces, forward slashes, and backslashes with underscores.
func RandomizeFileName(fileName string) string {
	// Extract the base file name (without the directory)
	baseName := filepath.Base(fileName)

	// Split the base name and extension
	extension := filepath.Ext(baseName)
	name := strings.TrimSuffix(baseName, extension)

	// Replace spaces and slashes with underscores
	name = strings.NewReplacer(" ", "_", "/", "_", "\\", "_").Replace(name)

	// Add the current timestamp to the file name
	now := strconv.FormatInt(time.Now().UnixNano(), 10)
	name = now + "_" + name

	return name + extension
}

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

// GetContentTypeFromExtension guesses the MIME type based on the file extension.
func GetContentTypeFromExtension(fileName string) string {
	mimeType := mime.TypeByExtension(filepath.Ext(fileName))
	if mimeType == "" {
		return "application/octet-stream" // Default MIME type
	}
	return mimeType
}

func ValidateContentType(contentType string, supportedMimeTypes []string) (bool, error) {
	for _, supportedType := range supportedMimeTypes {
		if supportedType == "*" || supportedType == "*/*" {
			return true, nil
		}

		if strings.HasSuffix(supportedType, "/*") {
			// Check if the MIME type matches the prefix (e.g., "image/*" matches "image/png")
			prefix := strings.TrimSuffix(supportedType, "/*")
			if strings.HasPrefix(contentType, prefix) {
				return true, nil
			}
		} else if contentType == supportedType {
			return true, nil
		}
	}
	return false, nil
}
