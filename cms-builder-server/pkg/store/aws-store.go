package store

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/clients"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	"github.com/rs/zerolog/log"
)

type S3Store struct {
	Client    *clients.AwsManager
	Config    *StoreConfig
	AwsConfig *S3Config
}

func (s *S3Store) GetConfig() *StoreConfig {
	return s.Config
}

func (s *S3Store) GetPath() string {
	return s.AwsConfig.Folder
}

// StoreFile uploads the given file to an S3 bucket using the provided configuration.
// It reads the file bytes and calls the UploadFile method of the AwsManager client.
// If successful, it returns a FileData object containing the file's name, path, and URL.
// If there is an error at any step, it logs the error and returns the error.
func (s *S3Store) StoreFile(fileName string, file multipart.File, header *multipart.FileHeader) (fileData *models.File, err error) {

	contentType := header.Header.Get("Content-Type")

	validContentType, err := ValidateContentType(contentType, s.Config.SupportedMimeTypes)
	if err != nil {
		return fileData, err
	}

	if !validContentType {
		return fileData, fmt.Errorf("invalid content type: %s", contentType)
	}

	fileBytes, err := getFileBytes(file)
	if err != nil {
		log.Error().Err(err).Msg("Error getting file bytes")
		return fileData, err
	}

	fileName = RandomizeFileName(fileName)

	// Create the uploads directory if it doesn't exist
	uploadsDir := s.GetPath()
	location, err := s.Client.UploadFile(uploadsDir, fileName, fileBytes)
	if err != nil {
		log.Error().Err(err).Msg("Error uploading file to S3")
		return fileData, err
	}

	url := "https://" + s.AwsConfig.Bucket + location

	fileData = &models.File{
		Name: fileName,
		Path: location,
		Url:  url,
	}

	fileInfo, err := s.GetFileInfo(fileData)
	if err != nil {
		return fileData, err
	}

	fileData.Size = fileInfo.Size
	fileData.MimeType = fileInfo.ContentType

	return fileData, nil
}

// DeleteFile deletes a file from the S3 bucket using the provided file path.
// It calls the DeleteFile method of the AwsManager client.
// If an error occurs during the deletion, it logs the error and returns it.
func (s *S3Store) DeleteFile(file models.File) error {
	log.Warn().Interface("file", file).Msg("Deleting file from S3")
	err := s.Client.DeleteFile(file.Path)
	if err != nil {
		log.Error().Err(err).Msg("Error deleting file from S3")
		return err
	}

	return nil
}

func (s *S3Store) ListFiles() ([]string, error) {
	return s.Client.ListFiles()
}

func (s *S3Store) ReadFile(file *models.File) ([]byte, error) {
	return s.Client.DownloadFile(file.Path)
}

func (s *S3Store) GetFileInfo(file *models.File) (*models.FileInfo, error) {
	return s.Client.GetFileInfo(file.Path)
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

type S3Config struct {
	Bucket    string
	Region    string
	AccessKey string
	SecretKey string
	Folder    string
}

// NewS3Store creates a new S3Store, which is used to store files in an AWS S3 bucket.
// It returns an error if the AWS configuration is not ready.
func NewS3Store(config *StoreConfig, awsConfig *S3Config) (*S3Store, error) {

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

	if config.Folder == "" {
		return nil, fmt.Errorf("uploader folder is empty")
	}

	if config.MaxSize >= 0 {
		return nil, fmt.Errorf("invalid max size: %d", config.MaxSize)
	}

	if len(config.SupportedMimeTypes) == 0 {
		return nil, fmt.Errorf("supported mime types is empty")
	}

	client := clients.AwsManager{
		Bucket:    awsConfig.Bucket,
		Region:    awsConfig.Region,
		AccessKey: awsConfig.AccessKey,
		SecretKey: awsConfig.SecretKey,
	}

	if !client.IsReady() {
		return nil, fmt.Errorf("AWS not ready")
	}

	return &S3Store{
		Client:    &client,
		AwsConfig: awsConfig,
		Config:    config,
	}, nil
}
