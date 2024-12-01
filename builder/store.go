package builder

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

type StoreType string

const (
	StoreLocal StoreType = "local"
	StoreS3    StoreType = "s3"
)

type Store interface {
	StoreFile(cfg *UploaderConfig, fileName string, file multipart.File) (fileData FileData, err error)
	DeleteFile(path string) error
}

type LocalStore struct{}

// StoreFile stores the given file in the local file system. It returns the
// FileData for the stored file, including the path to the file and the URL
// at which the file can be accessed. If the file cannot be stored, an error
// is returned.
func (s *LocalStore) StoreFile(cfg *UploaderConfig, fileName string, file multipart.File) (fileData FileData, err error) {

	fileData = FileData{}

	// Create the uploads directory if it doesn't exist
	uploadsDir := defaultUploadFolder
	if cfg.Folder != "" {
		uploadsDir = cfg.Folder
	}

	log.Warn().Interface("config", cfg).Str("uploadsDir", uploadsDir).Msg("Storing file")

	err = os.MkdirAll(uploadsDir, os.ModePerm)
	if err != nil {
		log.Error().Err(err).Msg("Error creating uploads directory")
		return fileData, err
	}

	// Create the file path
	fileData.Name = fileName
	fileData.Path = filepath.Join(uploadsDir, fileData.Name)

	// Save the file to disk
	dst, err := os.Create(fileData.Path)
	if err != nil {
		log.Error().Err(err).Str("path", fileData.Path).Msg("Error creating file")
		return fileData, err
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		log.Error().Err(err).Str("path", fileData.Path).Msg("Error saving file")
		return fileData, err
	}

	fileData.Url = filepath.Join(config.GetString(EnvKeys.BaseUrl), cfg.StaticPath, fileData.Name)

	return fileData, nil
}

// DeleteFile takes a file path and deletes the file from disk.
// It returns an error if the file cannot be deleted.
func (s *LocalStore) DeleteFile(path string) error {
	// Log the file path to be deleted
	log.Warn().Msgf("Deleting file: %s", path)

	// Attempt to delete the file
	if err := os.Remove(path); err != nil {
		// Log the error if the file cannot be deleted
		log.Println("Error deleting file:", err)
		return err
	}

	// Return nil if the file is successfully deleted
	return nil
}

type S3Store struct {
	Client *AwsManager
}

// StoreFile uploads the given file to an S3 bucket using the provided configuration.
// It reads the file bytes and calls the UploadFile method of the AwsManager client.
// If successful, it returns a FileData object containing the file's name, path, and URL.
// If there is an error at any step, it logs the error and returns the error.
func (s *S3Store) StoreFile(cfg *UploaderConfig, fileName string, file multipart.File) (fileData FileData, err error) {
	fileBytes, err := getFileBytes(file)
	if err != nil {
		log.Error().Err(err).Msg("Error getting file bytes")
		return fileData, err
	}

	// Create the uploads directory if it doesn't exist
	uploadsDir := defaultUploadFolder
	if cfg.Folder != "" {
		uploadsDir = cfg.Folder
	}

	err = s.Client.UploadFile(uploadsDir, fileName, fileBytes)
	if err != nil {
		log.Error().Err(err).Msg("Error uploading file to S3")
		return fileData, err
	}

	path := filepath.Join(uploadsDir, fileName)
	url := "https://" + config.GetString(EnvKeys.AwsBucket) + "/" + path

	fileData = FileData{
		Name: fileName,
		Path: path,
		Url:  url,
	}

	return fileData, nil
}

func getFileBytes(file multipart.File) ([]byte, error) {
	defer file.Close()

	var buf bytes.Buffer
	_, err := io.Copy(&buf, file)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (s *S3Store) DeleteFile(path string) error {
	err := s.Client.DeleteFile(path)
	if err != nil {
		log.Error().Err(err).Msg("Error deleting file from S3")
		return err
	}

	return nil
}

func NewS3Store() (*S3Store, error) {
	client := AwsManager{
		Bucket: config.GetString(EnvKeys.AwsBucket),
	}

	if !client.IsReady() {
		return nil, fmt.Errorf("AWS not ready")
	}

	return &S3Store{
		Client: &client,
	}, nil
}
