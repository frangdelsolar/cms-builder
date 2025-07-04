package store

import (
	"fmt"
	"mime"
	"mime/multipart"
	"net/http"
	"path/filepath"

	cliPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/clients"
	fileModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/file/models"
	fileTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/file/types"
	loggerTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger/types"
	storeTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/store/types"
	storeUtils "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/store/utils"
)

type FilebaseStore struct {
	Client         *cliPkg.FilebaseManager
	Config         *storeTypes.StoreConfig
	FilebaseConfig *storeTypes.S3Config
}

func (s *FilebaseStore) GetConfig() *storeTypes.StoreConfig {
	return s.Config
}

func (s *FilebaseStore) GetPath(file *fileModels.File) string {
	return s.FilebaseConfig.Folder + "/" + file.Name
}

// StoreFile uploads the given file to Filebase using the provided configuration.
// It reads the file bytes and calls the UploadFile method of the FilebaseManager client.
// If successful, it returns a FileData object containing the file's name, path, and URL.
// If there is an error at any step, it logs the error and returns the error.
func (s *FilebaseStore) StoreFile(fileName string, file multipart.File, header *multipart.FileHeader, log *loggerTypes.Logger) (fileData *fileModels.File, err error) {
	fileData = &fileModels.File{}

	fileBytes, err := getFileBytes(file)
	if err != nil {
		log.Error().Err(err).Msg("Error getting file bytes")
		return fileData, err
	}

	if len(fileBytes) == 0 {
		return fileData, fmt.Errorf("file is empty")
	}

	if len(fileBytes) > int(s.GetConfig().MaxSize) {
		return fileData, fmt.Errorf("file is too large")
	}

	// Detect the content type from the file content
	contentType := http.DetectContentType(fileBytes)

	log.Debug().Str("detected-content-type", contentType).Msg("Detected file content type")

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
	valid, err := storeUtils.ValidateContentType(contentType, s.Config.SupportedMimeTypes)
	if err != nil {
		return fileData, err
	}
	if !valid {
		return fileData, fmt.Errorf("invalid content type: %s", contentType)
	}

	fileData.Name = storeUtils.RandomizeFileName(fileName)

	// Create the uploads directory if it doesn't exist
	path := s.GetPath(fileData)
	location, err := s.Client.UploadFile(path, fileBytes, log)
	if err != nil {
		log.Error().Err(err).Msg("Error uploading file to Filebase")
		return fileData, err
	}

	url := "https://" + s.FilebaseConfig.Bucket + ".s3.filebase.com/" + location

	fileData.Path = location
	fileData.Url = url
	fileData.Size = int64(len(fileBytes))
	fileData.MimeType = contentType

	log.Info().
		Str("name", fileData.Name).
		Str("path", fileData.Path).
		Str("url", fileData.Url).
		Int64("size", fileData.Size).
		Str("mime-type", fileData.MimeType).
		Msg("File stored successfully in Filebase")

	return fileData, nil
}

// DeleteFile deletes a file from Filebase using the provided file path.
// It calls the DeleteFile method of the FilebaseManager client.
// If an error occurs during the deletion, it logs the error and returns it.
func (s *FilebaseStore) DeleteFile(file *fileModels.File, log *loggerTypes.Logger) error {
	log.Warn().Interface("file", file).Msg("Deleting file from Filebase")
	err := s.Client.DeleteFile(file.Path, log)
	if err != nil {
		log.Error().Err(err).Msg("Error deleting file from Filebase")
		return err
	}

	return nil
}

func (s *FilebaseStore) ListFiles(log *loggerTypes.Logger) ([]string, error) {
	return s.Client.ListFiles(log)
}

func (s *FilebaseStore) ReadFile(file *fileModels.File, log *loggerTypes.Logger) ([]byte, error) {
	return s.Client.DownloadFile(file.Path, log)
}

func (s *FilebaseStore) GetFileInfo(file *fileModels.File, log *loggerTypes.Logger) (*fileTypes.FileInfo, error) {
	info := &fileTypes.FileInfo{
		Name:        file.Name,
		Size:        file.Size,
		ContentType: file.MimeType,
	}

	return info, nil
}

// NewFilebaseStore creates a new FilebaseStore, which is used to store files in Filebase.
// It returns an error if the Filebase configuration is not ready.
func NewFilebaseStore(config *storeTypes.StoreConfig, filebaseConfig *storeTypes.S3Config) (*FilebaseStore, error) {
	if filebaseConfig == nil {
		return nil, fmt.Errorf("config is nil")
	}

	if filebaseConfig.Bucket == "" {
		return nil, fmt.Errorf("bucket is empty")
	}

	if filebaseConfig.Endpoint == "" {
		return nil, fmt.Errorf("endpoint is empty")
	}

	if filebaseConfig.AccessKey == "" {
		return nil, fmt.Errorf("access key is empty")
	}

	if filebaseConfig.SecretKey == "" {
		return nil, fmt.Errorf("secret key is empty")
	}

	if filebaseConfig.Folder == "" {
		return nil, fmt.Errorf("folder is empty")
	}

	if config == nil {
		return nil, fmt.Errorf("config is nil")
	}

	if config.MediaFolder == "" {
		return nil, fmt.Errorf("uploader folder is empty")
	}

	if config.MaxSize <= 0 {
		return nil, fmt.Errorf("invalid max size: %d", config.MaxSize)
	}

	if len(config.SupportedMimeTypes) == 0 {
		return nil, fmt.Errorf("supported mime types is empty")
	}

	client := cliPkg.FilebaseManager{
		Bucket:    filebaseConfig.Bucket,
		Endpoint:  filebaseConfig.Endpoint,
		Region:    filebaseConfig.Region,
		AccessKey: filebaseConfig.AccessKey,
		SecretKey: filebaseConfig.SecretKey,
	}

	return &FilebaseStore{
		Client:         &client,
		FilebaseConfig: filebaseConfig,
		Config:         config,
	}, nil
}
