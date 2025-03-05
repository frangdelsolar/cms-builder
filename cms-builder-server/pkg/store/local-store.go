package store

import (
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	loggerTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger/types"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
)

type LocalStore struct {
	MediaFolderAbsolutePath string // i. e. /Users/user/.../media/project-name
	BaseUrl                 string // baseUrl where files will be served
	Config                  *StoreConfig
}

func (s LocalStore) GetConfig() *StoreConfig {
	return s.Config
}

func NewLocalStore(config *StoreConfig, folder string, baseUrl string) (LocalStore, error) {
	if config == nil {
		return LocalStore{}, fmt.Errorf("config is nil")
	}

	if config.MediaFolder == "" {
		return LocalStore{}, fmt.Errorf("uploader folder is empty")
	}

	if config.MaxSize <= 0 {
		return LocalStore{}, fmt.Errorf("invalid max size: %d", config.MaxSize)
	}

	if len(config.SupportedMimeTypes) == 0 {
		return LocalStore{}, fmt.Errorf("supported mime types is empty")
	}

	if _, err := os.Stat(folder); os.IsNotExist(err) {
		err = os.MkdirAll(folder, os.ModePerm)
		if err != nil {
			return LocalStore{}, fmt.Errorf("failed to create directory: %v", err)
		}
	}

	absPath, err := filepath.Abs(folder)
	if err != nil {
		return LocalStore{}, fmt.Errorf("failed to get absolute path: %v", err)
	}

	return LocalStore{
		Config:                  config,
		MediaFolderAbsolutePath: absPath,
		BaseUrl:                 baseUrl,
	}, nil
}

func (s LocalStore) GetPath(file *models.File) string {
	return s.MediaFolderAbsolutePath + "/" + file.Name
}

func (s LocalStore) StoreFile(fileName string, file multipart.File, header *multipart.FileHeader, log *loggerTypes.Logger) (fileData *models.File, err error) {
	fileData = &models.File{}

	// make sure files is not empty
	if header.Size == 0 {
		return fileData, fmt.Errorf("file is empty")
	}

	if header.Size > int64(s.GetConfig().MaxSize) {
		return fileData, fmt.Errorf("file is too large")
	}

	// Read the first 512 bytes of the file to detect its MIME type
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil && err != io.EOF {
		return fileData, fmt.Errorf("failed to read file content: %v", err)
	}

	// Detect the content type from the file content
	contentType := http.DetectContentType(buffer)

	// If the detected content type is "application/octet-stream",
	// fall back to the file extension for a better guess
	if contentType == "application/octet-stream" {
		contentType = mime.TypeByExtension(filepath.Ext(fileName))
		if contentType == "" {
			contentType = "application/octet-stream" // Default MIME type
		}
	}

	// Validate the detected content type
	valid, err := ValidateContentType(contentType, s.Config.SupportedMimeTypes)
	if err != nil {
		return fileData, err
	}
	if !valid {
		return fileData, fmt.Errorf("invalid content type: %s", contentType)
	}

	// Create the file path
	fileData.Name = RandomizeFileName(fileName)
	path := s.GetPath(fileData)

	// Save the file to disk
	dst, err := os.Create(path)
	if err != nil {
		return fileData, fmt.Errorf("failed to create file: %v", err)
	}
	defer dst.Close()

	// Write the buffer (first 512 bytes) to the file
	_, err = dst.Write(buffer)
	if err != nil {
		return fileData, fmt.Errorf("failed to write file content: %v", err)
	}

	// Copy the remaining file content
	_, err = io.Copy(dst, file)
	if err != nil {
		return fileData, fmt.Errorf("failed to copy file content: %v", err)
	}

	// Get file info (size and MIME type)
	fileInfo, err := s.GetFileInfo(fileData, log)
	if err != nil {
		return fileData, err
	}

	fileData.Size = fileInfo.Size
	fileData.MimeType = fileInfo.ContentType
	fileData.Path = s.Config.MediaFolder + "/" + fileData.Name
	fileData.Url = "" // /private/api/files/{id}/download

	// Log the successful file upload
	log.Info().
		Str("name", fileData.Name).
		Int64("size", fileData.Size).
		Str("mime-type", fileData.MimeType).
		Msg("File stored successfully")

	return fileData, nil
}

// DeleteFile takes a file path and deletes the file from disk.
// It returns an error if the file cannot be deleted.
func (s LocalStore) DeleteFile(file *models.File, log *loggerTypes.Logger) error {
	path := s.GetPath(file)

	// Attempt to delete the file
	if err := os.Remove(path); err != nil {
		return err
	}

	// Return nil if the file is successfully deleted
	return nil
}

func (s LocalStore) ListFiles(log *loggerTypes.Logger) ([]string, error) {
	output := []string{}

	err := filepath.Walk(s.MediaFolderAbsolutePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err // Return the error to stop walking if needed
		}

		// Get the relative path from the store's path
		relPath, err := filepath.Rel(s.MediaFolderAbsolutePath, path)
		if err != nil {
			return err
		}

		// Skip the root directory itself.  Important!
		if relPath == "." {
			return nil // Continue walking
		}
		relPath = s.MediaFolderAbsolutePath + "/" + relPath

		output = append(output, relPath)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}

func (s LocalStore) ReadFile(file *models.File, log *loggerTypes.Logger) ([]byte, error) {
	path := s.GetPath(file)
	return os.ReadFile(path)
}

func (s LocalStore) GetFileInfo(file *models.File, log *loggerTypes.Logger) (*models.FileInfo, error) {
	path := s.GetPath(file)
	stats, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if stats.IsDir() {
		return nil, fmt.Errorf("file is a directory")
	}

	// Detect the content type using the utility function
	contentType, err := DetectContentTypeFromFile(path)
	if err != nil {
		return nil, err
	}

	fileInfo := &models.FileInfo{
		Name:         stats.Name(),
		Size:         stats.Size(),
		LastModified: stats.ModTime(),
		ContentType:  contentType,
	}

	return fileInfo, nil
}
