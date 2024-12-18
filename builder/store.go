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
	GetPath() string
	StoreFile(cfg *UploaderConfig, fileName string, file multipart.File) (fileData FileData, err error)
	DeleteFile(file FileData) error
	ListFiles() ([]string, error)
	ReadFile(file *FileData) ([]byte, error)
}

type LocalStore struct {
	Path string
}

func NewLocalStore(folder string) *LocalStore {
	return &LocalStore{
		Path: folder,
	}
}

func (s *LocalStore) GetPath() string {
	return s.Path
}

// StoreFile stores the given file in the local file system. It returns the
// FileData for the stored file, including the path to the file and the URL
// at which the file can be accessed. If the file cannot be stored, an error
// is returned.
func (s *LocalStore) StoreFile(cfg *UploaderConfig, fileName string, file multipart.File) (fileData FileData, err error) {

	fileData = FileData{}

	// Create the uploads directory if it doesn't exist
	uploadsDir := s.GetPath()

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

	fileData.Url = filepath.Join(config.GetString(EnvKeys.BaseUrl), "file", fileData.Name)

	return fileData, nil
}

// DeleteFile takes a file path and deletes the file from disk.
// It returns an error if the file cannot be deleted.
func (s *LocalStore) DeleteFile(file FileData) error {
	// Log the file path to be deleted
	log.Warn().Msgf("Deleting file: %s", file.Path)

	// Attempt to delete the file
	if err := os.Remove(file.Path); err != nil {
		// Log the error if the file cannot be deleted
		log.Println("Error deleting file:", err)
		return err
	}

	// Return nil if the file is successfully deleted
	return nil
}

func (s *LocalStore) ListFiles() ([]string, error) {
	log.Warn().Msgf("Listing files from %s", s.Path)
	output := []string{}
	files, err := os.ReadDir(s.Path)
	if err != nil {
		log.Error().Err(err).Msg("Error listing files")
		return output, err
	}

	for _, file := range files {
		output = append(output, file.Name())
	}
	return output, nil
}

func (s *LocalStore) ReadFile(file *FileData) ([]byte, error) {
	return os.ReadFile(file.Path)
}

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
func (s *S3Store) StoreFile(cfg *UploaderConfig, fileName string, file multipart.File) (fileData FileData, err error) {
	fileBytes, err := getFileBytes(file)
	if err != nil {
		log.Error().Err(err).Msg("Error getting file bytes")
		return fileData, err
	}

	// Create the uploads directory if it doesn't exist
	uploadsDir := s.GetPath()
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

// DeleteFile deletes a file from the S3 bucket using the provided file path.
// It calls the DeleteFile method of the AwsManager client.
// If an error occurs during the deletion, it logs the error and returns it.
func (s *S3Store) DeleteFile(file FileData) error {
	log.Warn().Interface("file", file).Msg("Deleting file from S3")
	err := s.Client.DeleteFile(file.Url)
	if err != nil {
		log.Error().Err(err).Msg("Error deleting file from S3")
		return err
	}

	return nil
}

func (s *S3Store) ListFiles() ([]string, error) {
	return s.Client.ListFiles()
}

func (s *S3Store) ReadFile(file *FileData) ([]byte, error) {
	return s.Client.DownloadFile(file.Path)
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
	client := AwsManager{
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
