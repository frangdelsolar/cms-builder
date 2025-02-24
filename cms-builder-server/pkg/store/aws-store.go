package store

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/clients"
	"github.com/rs/zerolog/log"
)

type S3Store struct {
	Client *AwsManager
	Path   string
}

func (s *S3Store) GetPath() string {
	return s.Path
}

// StoreFile uploads the given file to an S3 bucket using the provided configuration.
// It reads the file bytes and calls the UploadFile method of the AwsManager client.
// If successful, it returns a FileData object containing the file's name, path, and URL.
// If there is an error at any step, it logs the error and returns the error.
func (s *S3Store) StoreFile(cfg *StoreConfig, fileName string, file multipart.File) (fileData File, err error) {
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

	url := "https://" + config.GetString(EnvKeys.AwsBucket) + location

	fileData = File{
		Name: fileName,
		Path: location,
		Url:  url,
	}

	return fileData, nil
}

// DeleteFile deletes a file from the S3 bucket using the provided file path.
// It calls the DeleteFile method of the AwsManager client.
// If an error occurs during the deletion, it logs the error and returns it.
func (s *S3Store) DeleteFile(file File) error {
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

func (s *S3Store) ReadFile(file *File) ([]byte, error) {
	return s.Client.DownloadFile(file.Path)
}

func (s *S3Store) GetFileInfo(file *File) (*FileInfo, error) {
	return s.Client.GetFileInfo(file.Path)
}

// getFileBytes reads the contents of a multipart.File into a byte array.
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
func NewS3Store(folder string) (*S3Store, error) {
	client := clients.AwsManager{
		Bucket: config.GetString(EnvKeys.AwsBucket),
	}

	if !client.IsReady() {
		return nil, fmt.Errorf("AWS not ready")
	}

	return &S3Store{
		Client: &client,
		Path:   folder,
	}, nil
}
