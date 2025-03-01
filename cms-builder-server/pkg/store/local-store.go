package store

import (
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
)

type LocalStore struct {
	Path    string
	BaseUrl string
	Config  *StoreConfig
}

func (s *LocalStore) GetConfig() *StoreConfig {
	return s.Config
}

func NewLocalStore(config *StoreConfig, folder string, baseUrl string) (*LocalStore, error) {
	if config == nil {
		return nil, fmt.Errorf("config is nil")
	}

	if config.Folder == "" {
		return nil, fmt.Errorf("uploader folder is empty")
	}

	if config.MaxSize <= 0 {
		return nil, fmt.Errorf("invalid max size: %d", config.MaxSize)
	}

	if len(config.SupportedMimeTypes) == 0 {
		return nil, fmt.Errorf("supported mime types is empty")
	}

	dir := "./" + folder
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.Mkdir(dir, os.ModePerm)
		if err != nil {
			return nil, fmt.Errorf("failed to create directory: %v", err)
		}
	}

	absPath, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %v", err)
	}

	// trim start slash
	if absPath[0] == '/' {
		absPath = absPath[1:]
	}

	return &LocalStore{
		Config:  config,
		Path:    absPath,
		BaseUrl: baseUrl,
	}, nil
}

func (s *LocalStore) GetPath() string {
	return s.Path
}

func (s *LocalStore) StoreFile(fileName string, file multipart.File, header *multipart.FileHeader, log *logger.Logger) (fileData *models.File, err error) {
	fileData = &models.File{}

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

	log.Debug().Str("detected-content-type", contentType).Msg("Detected file content type")

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
	fileData.Path = filepath.Join(s.Path, fileData.Name)

	// Save the file to disk
	dst, err := os.Create(fileData.Path)
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

	// Set the file URL
	fileData.Url = s.BaseUrl + "/static/" + fileData.Name

	// Get file info (size and MIME type)
	fileInfo, err := s.GetFileInfo(fileData, log)
	if err != nil {
		return fileData, err
	}

	fileData.Size = fileInfo.Size
	fileData.MimeType = fileInfo.ContentType

	// Log the successful file upload
	log.Info().
		Str("name", fileData.Name).
		Str("path", fileData.Path).
		Str("url", fileData.Url).
		Int64("size", fileData.Size).
		Str("mime-type", fileData.MimeType).
		Msg("File stored successfully")

	return fileData, nil
}

// DeleteFile takes a file path and deletes the file from disk.
// It returns an error if the file cannot be deleted.
func (s *LocalStore) DeleteFile(file *models.File, log *logger.Logger) error {

	log.Info().Str("path", file.Path).Msg("Deleting file")

	// Attempt to delete the file
	if err := os.Remove(file.Path); err != nil {
		return err
	}

	// Return nil if the file is successfully deleted
	return nil
}

func (s *LocalStore) ListFiles(log *logger.Logger) ([]string, error) {
	output := []string{}

	err := filepath.Walk(s.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err // Return the error to stop walking if needed
		}

		// Get the relative path from the store's path
		relPath, err := filepath.Rel(s.Path, path)
		if err != nil {
			return err
		}

		// Skip the root directory itself.  Important!
		if relPath == "." {
			return nil // Continue walking
		}
		relPath = s.Path + "/" + relPath

		output = append(output, relPath)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}

func (s *LocalStore) ReadFile(file *models.File, log *logger.Logger) ([]byte, error) {
	return os.ReadFile(file.Path)
}

func (s *LocalStore) GetFileInfo(file *models.File, log *logger.Logger) (*models.FileInfo, error) {
	stats, err := os.Stat(file.Path)
	if err != nil {
		return nil, err
	}

	if stats.IsDir() {
		return nil, fmt.Errorf("file is a directory")
	}

	// Detect the content type using the utility function
	contentType, err := DetectContentTypeFromFile(file.Path)
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
