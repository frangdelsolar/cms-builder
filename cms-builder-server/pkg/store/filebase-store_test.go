package store_test

import (
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	fileModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/file/models"
	loggerPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger"
	storePkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/store"
	storeTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/store/types"
)

// Helper function to create a new Filebase store for testing
func createFilebaseStore(t *testing.T) *storePkg.FilebaseStore {
	if os.Getenv("AWS_BUCKET") == "" {
		godotenv.Load(".test.env")
	}

	config := &storeTypes.StoreConfig{
		MediaFolder:        "test_media",
		MaxSize:            1024 * 1024, // 1MB
		SupportedMimeTypes: []string{"image/jpeg", "image/png"},
	}

	filebaseConfig := &storeTypes.S3Config{
		Bucket:    os.Getenv("AWS_BUCKET"),
		Region:    os.Getenv("AWS_REGION"),
		AccessKey: os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
		Endpoint:  os.Getenv("AWS_ENDPOINT"),
		Folder:    config.MediaFolder,
	}

	store, err := storePkg.NewFilebaseStore(config, filebaseConfig)
	require.NoError(t, err)
	return store
}

// Helper function to create a test file in Filebase
func createFilebaseTestFile(t *testing.T, store *storePkg.FilebaseStore, fileName string, content []byte) *fileModels.File {
	file, fileHeader := createMultipartFile(t, fileName, content)

	fileData, err := store.StoreFile(fileHeader.Filename, file, fileHeader, loggerPkg.Default)
	assert.NoError(t, err)
	assert.NotNil(t, fileData)

	return fileData
}

// Helper function to clean up test files from Filebase
func cleanupFilebaseTestMedia(t *testing.T, store *storePkg.FilebaseStore, file *fileModels.File) {
	log := loggerPkg.Default
	err := store.DeleteFile(file, log)
	assert.NoError(t, err)
}

func TestNewFilebaseStore(t *testing.T) {
	// Test valid configuration
	t.Run("Valid Configuration", func(t *testing.T) {
		store := createFilebaseStore(t)
		assert.NotNil(t, store)
	})

	// Test invalid configuration
	t.Run("Missing Endpoint", func(t *testing.T) {
		config := &storeTypes.StoreConfig{
			MediaFolder:        "test_media",
			MaxSize:            1024 * 1024,
			SupportedMimeTypes: []string{"image/jpeg"},
		}

		invalidConfig := &storeTypes.S3Config{
			Bucket:    "test-bucket-ef",
			Region:    "us-east-1",
			AccessKey: "access-key",
			SecretKey: "secret-key",
			Folder:    "test_media",
			// Missing Endpoint
		}

		_, err := storePkg.NewFilebaseStore(config, invalidConfig)
		assert.Error(t, err)
	})
}

func TestFilebaseStoreFile(t *testing.T) {
	store := createFilebaseStore(t)

	// Test storing a file
	t.Run("Store JPEG File Successfully", func(t *testing.T) {
		fileData := createFilebaseTestFile(t, store, "testfile.jpg", jpegFileContent)
		defer cleanupFilebaseTestMedia(t, store, fileData)

		t.Log(fileData)

		assert.Contains(t, fileData.Name, "testfile.jpg")
		assert.Equal(t, "image/jpeg", fileData.MimeType)
		assert.Equal(t, int64(len(jpegFileContent)), fileData.Size)
	})

	// Test storing an unsupported file type
	t.Run("Unsupported File Type", func(t *testing.T) {
		unsupportedFileContent := []byte("unsupported file content")
		unsupportedFile, unsupportedFileHeader := createMultipartFile(t, "testfile.txt", unsupportedFileContent)

		log := loggerPkg.Default
		_, err := store.StoreFile(unsupportedFileHeader.Filename, unsupportedFile, unsupportedFileHeader, log)
		assert.Error(t, err)
	})

	// Test storing a file that's too large
	t.Run("File Too Large", func(t *testing.T) {
		// Create a file larger than the max size (1MB)
		largeFileContent := make([]byte, 2*1024*1024) // 2MB
		largeFile, largeFileHeader := createMultipartFile(t, "largefile.jpg", largeFileContent)

		log := loggerPkg.Default
		_, err := store.StoreFile(largeFileHeader.Filename, largeFile, largeFileHeader, log)
		assert.Error(t, err)
	})
}

func TestFilebaseDeleteFile(t *testing.T) {
	store := createFilebaseStore(t)

	// Create a test file
	fileData := createFilebaseTestFile(t, store, "testfile.jpg", jpegFileContent)

	// Test deleting a file
	t.Run("Delete File Successfully", func(t *testing.T) {
		log := loggerPkg.Default
		err := store.DeleteFile(fileData, log)
		assert.NoError(t, err)

		// Verify file is really gone
		_, err = store.ReadFile(fileData, log)
		assert.Error(t, err)
	})

	// Test deleting non-existent file
	// t.Run("Delete Non-Existent File", func(t *testing.T) {
	// 	log := loggerPkg.Default
	// 	nonExistentFile := &fileModels.File{
	// 		Path: "nonexistent.jpg",
	// 	}
	// 	err := store.DeleteFile(nonExistentFile, log)
	// 	assert.Error(t, err)
	// })
}

func TestFilebaseListFiles(t *testing.T) {
	store := createFilebaseStore(t)

	// Create test files
	fileNames := []string{"file1.jpg", "file2.jpg"}
	createdFiles := make([]fileModels.File, 0)
	for _, fileName := range fileNames {
		fileData := createFilebaseTestFile(t, store, fileName, jpegFileContent)
		createdFiles = append(createdFiles, *fileData)
	}

	// Test listing files
	t.Run("List Files Successfully", func(t *testing.T) {
		log := loggerPkg.Default
		listed, err := store.ListFiles(log)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(listed), len(createdFiles))

		// Verify our test files are in the list
		for _, file := range createdFiles {
			found := false
			for _, listedFile := range listed {
				if listedFile == file.Path {
					found = true
					break
				}
			}
			assert.True(t, found, "File %s not found in listing", file.Path)
		}
	})

	// Clean up
	for _, file := range createdFiles {
		cleanupFilebaseTestMedia(t, store, &file)
	}
}

func TestFilebaseReadFile(t *testing.T) {
	store := createFilebaseStore(t)

	// Create a test file
	fileData := createFilebaseTestFile(t, store, "testfile.jpg", jpegFileContent)
	defer cleanupFilebaseTestMedia(t, store, fileData)

	// Test reading a file
	t.Run("Read File Successfully", func(t *testing.T) {
		log := loggerPkg.Default
		content, err := store.ReadFile(fileData, log)
		require.NoError(t, err)
		assert.Equal(t, jpegFileContent, content)
	})

	// Test reading non-existent file
	t.Run("Read Non-Existent File", func(t *testing.T) {
		log := loggerPkg.Default
		nonExistentFile := &fileModels.File{
			Path: "nonexistent.jpg",
		}
		_, err := store.ReadFile(nonExistentFile, log)
		assert.Error(t, err)
	})
}

func TestFilebaseGetFileInfo(t *testing.T) {
	store := createFilebaseStore(t)

	// Create a test file
	fileData := createFilebaseTestFile(t, store, "testfile.jpg", jpegFileContent)
	defer cleanupFilebaseTestMedia(t, store, fileData)

	// Test getting file info
	t.Run("Get File Info Successfully", func(t *testing.T) {
		log := loggerPkg.Default
		fileInfo, err := store.GetFileInfo(fileData, log)
		assert.NoError(t, err)
		assert.Equal(t, fileData.Name, fileInfo.Name)
		assert.Equal(t, int64(len(jpegFileContent)), fileInfo.Size)
		assert.Equal(t, "image/jpeg", fileInfo.ContentType)
	})

	// Test getting info for non-existent file
	// t.Run("Get Info for Non-Existent File", func(t *testing.T) {
	// 	log := loggerPkg.Default
	// 	nonExistentFile := &fileModels.File{
	// 		Path: "nonexistent.jpg",
	// 	}
	// 	_, err := store.GetFileInfo(nonExistentFile, log)
	// 	assert.Error(t, err)
	// })
}
