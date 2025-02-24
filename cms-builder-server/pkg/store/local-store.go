package store

// import (
// 	"fmt"
// 	"io"
// 	"mime"
// 	"mime/multipart"
// 	"os"
// 	"path/filepath"

// 	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
// )

// type LocalStore struct {
// 	Path string
// }

// func NewLocalStore(folder string) *LocalStore {
// 	return &LocalStore{
// 		Path: folder,
// 	}
// }

// func (s *LocalStore) GetPath() string {
// 	return s.Path
// }

// // StoreFile stores the given file in the local file system. It returns the
// // FileData for the stored file, including the path to the file and the URL
// // at which the file can be accessed. If the file cannot be stored, an error
// // is returned.
// func (s *LocalStore) StoreFile(cfg *StoreConfig, fileName string, file multipart.File) (fileData models.File, err error) {

// 	fileData = models.File{}

// 	// Create the uploads directory if it doesn't exist
// 	uploadsDir := s.GetPath()

// 	err = os.MkdirAll(uploadsDir, os.ModePerm)
// 	if err != nil {
// 		log.Error().Err(err).Msg("Error creating uploads directory")
// 		return fileData, err
// 	}

// 	// Create the file path
// 	fileData.Name = RandomizeFileName(fileName)
// 	fileData.Path = filepath.Join(uploadsDir, fileData.Name)

// 	// Save the file to disk
// 	dst, err := os.Create(fileData.Path)
// 	if err != nil {
// 		log.Error().Err(err).Str("path", fileData.Path).Msg("Error creating file")
// 		return fileData, err
// 	}
// 	defer dst.Close()

// 	_, err = io.Copy(dst, file)
// 	if err != nil {
// 		log.Error().Err(err).Str("path", fileData.Path).Msg("Error saving file")
// 		return fileData, err
// 	}

// 	fileData.Url = config.GetString(EnvKeys.BaseUrl) + "/static/" + fileData.Name

// 	return fileData, nil
// }

// // DeleteFile takes a file path and deletes the file from disk.
// // It returns an error if the file cannot be deleted.
// func (s *LocalStore) DeleteFile(file models.File) error {
// 	// Log the file path to be deleted
// 	log.Warn().Interface("file", file).Msg("Deleting file from local store")

// 	// Attempt to delete the file
// 	if err := os.Remove(file.Path); err != nil {
// 		// Log the error if the file cannot be deleted
// 		log.Println("Error deleting file:", err)
// 		return err
// 	}

// 	// Return nil if the file is successfully deleted
// 	return nil
// }

// func (s *LocalStore) ListFiles() ([]string, error) {
// 	output := []string{}

// 	err := filepath.Walk(s.Path, func(path string, info os.FileInfo, err error) error {
// 		if err != nil {
// 			log.Error().Err(err).Msgf("Error accessing path: %s", path)
// 			return err // Return the error to stop walking if needed
// 		}

// 		// Get the relative path from the store's path
// 		relPath, err := filepath.Rel(s.Path, path)
// 		if err != nil {
// 			log.Error().Err(err).Msgf("Error getting relative path for: %s", path)
// 			return err
// 		}

// 		// Skip the root directory itself.  Important!
// 		if relPath == "." {
// 			return nil // Continue walking
// 		}
// 		relPath = s.Path + "/" + relPath

// 		output = append(output, relPath)
// 		return nil
// 	})

// 	if err != nil {
// 		log.Error().Err(err).Msg("Error walking the file tree")
// 		return nil, err
// 	}

// 	return output, nil
// }

// func (s *LocalStore) ReadFile(file *models.File) ([]byte, error) {
// 	return os.ReadFile(file.Path)
// }

// func (s *LocalStore) GetFileInfo(file *models.File) (*FileInfo, error) {

// 	stats, err := os.Stat(file.Path)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if stats.IsDir() {
// 		return nil, fmt.Errorf("file is a directory")
// 	}

// 	// Get the content type
// 	var contentType string
// 	mime := mime.TypeByExtension(filepath.Ext(file.Name))
// 	if mime != "" {
// 		contentType = mime
// 	} else {
// 		contentType = "application/octet-stream"
// 	}

// 	fileInfo := &FileInfo{
// 		Name:         stats.Name(),
// 		Size:         stats.Size(),
// 		LastModified: stats.ModTime(),
// 		ContentType:  contentType,
// 	}

// 	return fileInfo, nil
// }
