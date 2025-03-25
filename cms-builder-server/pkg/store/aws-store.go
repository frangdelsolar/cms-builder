package store

import (
	"bytes"
	"fmt"
	"io"
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

type S3Store struct {
	Client    *cliPkg.AwsManager
	Config    *storeTypes.StoreConfig
	AwsConfig *storeTypes.S3Config
}

func (s *S3Store) GetConfig() *storeTypes.StoreConfig {
	return s.Config
}

func (s *S3Store) GetPath(file *fileModels.File) string {
	return s.AwsConfig.Folder + "/" + file.Name
}

// StoreFile uploads the given file to an S3 bucket using the provided configuration.
// It reads the file bytes and calls the UploadFile method of the AwsManager client.
// If successful, it returns a FileData object containing the file's name, path, and URL.
// If there is an error at any step, it logs the error and returns the error.
func (s *S3Store) StoreFile(fileName string, file multipart.File, header *multipart.FileHeader, log *loggerTypes.Logger) (fileData *fileModels.File, err error) {
	fileData = &fileModels.File{}

	fileBytes, err := getFileBytes(file)
	if err != nil {
		log.Error().Err(err).Msg("Error getting file bytes")
		return fileData, err
	}

	if len(fileBytes) == 0 {
		return fileData, fmt.Errorf("file is empty")
	}

	// TODO update this logic -> at least define what maxsize unit is
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
		log.Error().Err(err).Msg("Error uploading file to S3")
		return fileData, err
	}

	url := "https://" + s.AwsConfig.Bucket + "/" + location

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
		Msg("File stored successfully")

	return fileData, nil
}

// DeleteFile deletes a file from the S3 bucket using the provided file path.
// It calls the DeleteFile method of the AwsManager client.
// If an error occurs during the deletion, it logs the error and returns it.
func (s *S3Store) DeleteFile(file *fileModels.File, log *loggerTypes.Logger) error {
	log.Warn().Interface("file", file).Msg("Deleting file from S3")
	err := s.Client.DeleteFile(file.Path, log)
	if err != nil {
		log.Error().Err(err).Msg("Error deleting file from S3")
		return err
	}

	return nil
}

func (s *S3Store) ListFiles(log *loggerTypes.Logger) ([]string, error) {
	return s.Client.ListFiles(log)
}

func (s *S3Store) ReadFile(file *fileModels.File, log *loggerTypes.Logger) ([]byte, error) {
	return s.Client.DownloadFile(file.Path, log)
}

func (s *S3Store) GetFileInfo(file *fileModels.File, log *loggerTypes.Logger) (*fileTypes.FileInfo, error) {
	info := &fileTypes.FileInfo{
		Name:        file.Name,
		Size:        file.Size,
		ContentType: file.MimeType,
	}

	return info, nil
}

// getFileBytes reads the contents of a multipart.models.File into a byte array.
// It defers calling Close() on the file and returns an error if there is an
// error copying the file's contents.
func getFileBytes(file multipart.File) ([]byte, error) {
	defer file.Close()

	var buf bytes.Buffer
	_, err := io.Copy(&buf, file)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// NewS3Store creates a new S3Store, which is used to store files in an AWS S3 bucket.
// It returns an error if the AWS configuration is not ready.
func NewS3Store(config *storeTypes.StoreConfig, awsConfig *storeTypes.S3Config) (*S3Store, error) {

	if awsConfig == nil {
		return nil, fmt.Errorf("config is nil")
	}

	if awsConfig.Bucket == "" {
		return nil, fmt.Errorf("bucket is empty")
	}

	if awsConfig.Region == "" {
		return nil, fmt.Errorf("region is empty")
	}

	if awsConfig.AccessKey == "" {
		return nil, fmt.Errorf("access key is empty")
	}

	if awsConfig.SecretKey == "" {
		return nil, fmt.Errorf("secret key is empty")
	}

	if awsConfig.Folder == "" {
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

	client := cliPkg.AwsManager{
		Bucket:    awsConfig.Bucket,
		Region:    awsConfig.Region,
		AccessKey: awsConfig.AccessKey,
		SecretKey: awsConfig.SecretKey,
	}

	// FIXME
	// if !client.IsReady() {
	// 	return nil, fmt.Errorf("AWS not ready")
	// }

	return &S3Store{
		Client:    &client,
		AwsConfig: awsConfig,
		Config:    config,
	}, nil
}
