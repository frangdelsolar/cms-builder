package builder

import (
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type StoreType string

const (
	StoreLocal StoreType = "local"
	StoreS3    StoreType = "s3"
)

type Store interface {
	StoreFile(cfg *UploaderConfig, header *multipart.FileHeader, file multipart.File) (fileData FileData, err error)
	DeleteFile(path string) error
}

type LocalStore struct{}

func (s *LocalStore) StoreFile(cfg *UploaderConfig, header *multipart.FileHeader, file multipart.File) (fileData FileData, err error) {

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
	fileData.Name = randomizeFileName(header.Filename)
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

type S3Store struct{}

func (s *S3Store) StoreFile(cfg *UploaderConfig, header *multipart.FileHeader, file multipart.File) (fileData FileData, err error) {
	log.Info().Msg("Uploading file to S3. Not implemented yet")
	return fileData, nil
}
func (s *S3Store) DeleteFile(path string) error {
	// Log the file path to be deleted
	log.Warn().Msgf("Deleting file: %s. Not implemented yet", path)

	// Return nil if the file is successfully deleted
	return nil
}

// randomizeFileName takes a file name and returns a new file name that is
// randomized using the current timestamp. The file name is also sanitized
// to replace spaces, forward slashes, and backslashes with underscores.
func randomizeFileName(fileName string) string {
	// Replace spaces with underscores
	fileName = strings.ReplaceAll(fileName, " ", "_")

	// Replace forward slashes with underscores
	fileName = strings.ReplaceAll(fileName, "/", "_")

	// Replace backslashes with underscores
	fileName = strings.ReplaceAll(fileName, "\\", "_")

	// Add the current timestamp to the file name
	now := strconv.FormatInt(time.Now().UnixNano(), 10)
	fileName = now + "_" + fileName

	return fileName
}
