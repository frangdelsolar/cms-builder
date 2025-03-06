package store_test

import (
	"bytes"
	"mime/multipart"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/file/models"
	loggerPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/store"
	storePkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/store"
	storeTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/store/types"
)

// Helper type to wrap bytes.Reader and implement multipart.File
type bytesFile struct {
	*bytes.Reader
}

func (bf *bytesFile) Close() error {
	// No-op, since bytes.Reader doesn't need to be closed
	return nil
}

// Helper function to create a new LocalStore for testing
func createLocalStore(t *testing.T) storePkg.LocalStore {
	config := &storeTypes.StoreConfig{
		MediaFolder:        "test_media",
		MaxSize:            1024 * 1024, // 1MB
		SupportedMimeTypes: []string{"image/jpeg", "image/png"},
	}

	ls, err := storePkg.NewLocalStore(config, "test_media", "http://localhost:8080")
	require.NoError(t, err)
	return ls
}

// Helper function to create a test file in the test media folder
func createTestFile(t *testing.T, fileName string, content []byte) string {
	filePath := filepath.Join("test_media", fileName)
	err := os.WriteFile(filePath, content, 0644)
	require.NoError(t, err)
	return filePath
}

// Helper function to clean up the test media folder
func cleanupTestMedia(t *testing.T) {
	err := os.RemoveAll("test_media")
	require.NoError(t, err)
}

// Helper function to create a multipart file for testing
func createMultipartFile(t *testing.T, fileName string, content []byte) (*bytesFile, *multipart.FileHeader) {
	file := &bytesFile{bytes.NewReader(content)}
	fileHeader := &multipart.FileHeader{
		Filename: fileName,
		Size:     int64(len(content)),
	}
	return file, fileHeader
}

func TestNewLocalStore(t *testing.T) {
	// Test valid configuration
	t.Run("Valid Configuration", func(t *testing.T) {
		ls := createLocalStore(t)
		assert.NotNil(t, ls)
		assert.Contains(t, ls.MediaFolderAbsolutePath, "test_media")
	})

	// Test invalid configuration
	t.Run("Invalid Configuration", func(t *testing.T) {
		_, err := store.NewLocalStore(nil, "test_media", "http://localhost:8080")
		assert.Error(t, err)
	})

	// Clean up
	defer cleanupTestMedia(t)
}

func TestStoreFile(t *testing.T) {
	ls := createLocalStore(t)

	// Create a test file
	fileContent := []byte("test file content")
	file, fileHeader := createMultipartFile(t, "testfile.jpg", fileContent)

	// Test storing a file
	t.Run("Store File Successfully", func(t *testing.T) {
		log := loggerPkg.Default
		fileData, err := ls.StoreFile(fileHeader.Filename, file, fileHeader, log)
		assert.NoError(t, err)
		assert.NotNil(t, fileData)
		assert.Contains(t, fileData.Name, fileHeader.Filename)
	})

	// Test storing an unsupported file type
	t.Run("Unsupported File Type", func(t *testing.T) {
		unsupportedFileContent := []byte("unsupported file content")
		unsupportedFile, unsupportedFileHeader := createMultipartFile(t, "testfile.txt", unsupportedFileContent)

		log := loggerPkg.Default
		_, err := ls.StoreFile(unsupportedFileHeader.Filename, unsupportedFile, unsupportedFileHeader, log)
		assert.Error(t, err)
	})

	// Clean up
	defer cleanupTestMedia(t)
}

func TestDeleteFile(t *testing.T) {
	ls := createLocalStore(t)

	// Create a test file
	fileContent := []byte("test file content")
	filePath := createTestFile(t, "testfile.jpg", fileContent)

	// Test deleting a file
	t.Run("Delete File Successfully", func(t *testing.T) {
		log := loggerPkg.Default
		file := &models.File{
			Name: "testfile.jpg",
			Path: filePath,
		}
		err := ls.DeleteFile(file, log)
		assert.NoError(t, err)
		_, err = os.Stat(filePath)
		assert.True(t, os.IsNotExist(err))
	})

	// Clean up
	defer cleanupTestMedia(t)
}

func TestListFiles(t *testing.T) {
	ls := createLocalStore(t)

	// Create test files
	fileNames := []string{"file1.jpg", "file2.jpg"}
	for _, fileName := range fileNames {
		createTestFile(t, fileName, []byte("test content"))
	}

	// Test listing files
	t.Run("List Files Successfully", func(t *testing.T) {
		log := loggerPkg.Default
		files, err := ls.ListFiles(log)
		require.NoError(t, err)
		assert.Len(t, files, len(fileNames))
	})

	// Clean up
	defer cleanupTestMedia(t)
}

func TestReadFile(t *testing.T) {
	ls := createLocalStore(t)

	// Create a test file
	fileContent := []byte("test file content")
	filePath := createTestFile(t, "testfile.jpg", fileContent)

	// Test reading a file
	t.Run("Read File Successfully", func(t *testing.T) {
		log := loggerPkg.Default
		file := &models.File{
			Name: "testfile.jpg",
			Path: filePath,
		}
		content, err := ls.ReadFile(file, log)
		require.NoError(t, err)
		assert.Equal(t, fileContent, content)
	})

	// Clean up
	defer cleanupTestMedia(t)
}

func TestGetFileInfo(t *testing.T) {
	ls := createLocalStore(t)

	// Create a test file
	fileContent := []byte("test file content")
	filePath := createTestFile(t, "testfile.jpg", fileContent)

	// Test getting file info
	t.Run("Get File Info Successfully", func(t *testing.T) {
		log := loggerPkg.Default
		file := &models.File{
			Name: "testfile.jpg",
			Path: filePath,
		}
		fileInfo, err := ls.GetFileInfo(file, log)
		require.NoError(t, err)
		assert.Equal(t, file.Name, fileInfo.Name)
		assert.Equal(t, int64(len(fileContent)), fileInfo.Size)
		assert.Equal(t, "image/jpeg", fileInfo.ContentType)
	})

	// Clean up
	defer cleanupTestMedia(t)
}
