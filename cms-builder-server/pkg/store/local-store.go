package store

import (
	"fmt"
	"io"
	"mime"
	"mime/multipart"
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

	return &LocalStore{
		Config:  config,
		Path:    config.Folder,
		BaseUrl: baseUrl,
	}, nil
}

func (s *LocalStore) GetPath() string {
	return s.Path
}

// StoreFile stores the given file in the local file system. It returns the
// FileData for the stored file, including the path to the file and the URL
// at which the file can be accessed. If the file cannot be stored, an error
// is returned.
func (s *LocalStore) StoreFile(fileName string, file multipart.File, header *multipart.FileHeader, log *logger.Logger) (fileData *models.File, err error) {

	fileData = &models.File{}

	// Create the uploads directory if it doesn't exist
	uploadsDir := s.GetPath()

	err = os.MkdirAll(uploadsDir, os.ModePerm)
	if err != nil {
		return fileData, err
	}

	contentType := header.Header.Get("Content-Type")

	log.Debug().Str("content-type", contentType).Msg("File Content Type")

	validContentType, err := ValidateContentType(contentType, s.Config.SupportedMimeTypes)
	if err != nil {
		return fileData, err
	}

	if !validContentType {
		return fileData, fmt.Errorf("invalid content type: %s", contentType)
	}

	// Create the file path
	fileData.Name = RandomizeFileName(fileName)
	fileData.Path = filepath.Join(uploadsDir, fileData.Name)

	// Save the file to disk
	dst, err := os.Create(fileData.Path)
	if err != nil {
		return fileData, err
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		return fileData, err
	}

	fileData.Url = s.BaseUrl + "/static/" + fileData.Name

	fileInfo, err := s.GetFileInfo(fileData, log)
	if err != nil {
		return fileData, err
	}

	fileData.Size = fileInfo.Size
	fileData.MimeType = fileInfo.ContentType

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

	// Get the content type
	var contentType string
	mime := mime.TypeByExtension(filepath.Ext(file.Name))
	if mime != "" {
		contentType = mime
	} else {
		contentType = "application/octet-stream"
	}

	fileInfo := &models.FileInfo{
		Name:         stats.Name(),
		Size:         stats.Size(),
		LastModified: stats.ModTime(),
		ContentType:  contentType,
	}

	return fileInfo, nil
}
