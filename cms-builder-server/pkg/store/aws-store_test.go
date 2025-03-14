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

// Helper function to create a new S3 for testing
func createS3Store(t *testing.T) *storePkg.S3Store {

	if os.Getenv("AWS_BUCKET") == "" {
		godotenv.Load(".test.env")
	}

	config := &storeTypes.StoreConfig{
		MediaFolder:        "test_media",
		MaxSize:            1024 * 1024, // 1MB
		SupportedMimeTypes: []string{"image/jpeg", "image/png"},
	}

	awsConfig := &storeTypes.S3Config{
		Bucket:    os.Getenv("AWS_BUCKET"),
		Region:    os.Getenv("AWS_REGION"),
		AccessKey: os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
		Folder:    config.MediaFolder,
	}

	s3, err := storePkg.NewS3Store(config, awsConfig)
	require.NoError(t, err)
	return s3
}

// Helper function to create a test file in the test media folder
func createS3TestFile(t *testing.T, s3 *storePkg.S3Store, fileName string, content []byte) *fileModels.File {

	// Create a test file
	file, fileHeader := createMultipartFile(t, fileName, content)

	fileData, err := s3.StoreFile(fileHeader.Filename, file, fileHeader, loggerPkg.Default)
	assert.NoError(t, err)
	assert.NotNil(t, fileData)

	return fileData
}

// Helper function to clean up the test media folder
func cleanupS3TestMedia(t *testing.T, s3 *storePkg.S3Store, file *fileModels.File) {
	log := loggerPkg.Default
	err := s3.Client.DeleteFile(file.Path, log)
	assert.NoError(t, err)
}

func TestNewS3Store(t *testing.T) {
	// Test valid configuration
	t.Run("Valid Configuration", func(t *testing.T) {
		s3 := createS3Store(t)
		assert.NotNil(t, s3)
		assert.Contains(t, s3.AwsConfig.Folder, "test_media")
	})

	// Test invalid configuration
	t.Run("Invalid Configuration", func(t *testing.T) {
		_, err := storePkg.NewS3Store(nil, nil)
		assert.Error(t, err)
	})

}

func TestS3StoreFile(t *testing.T) {
	s3 := createS3Store(t)

	// Create a test file
	file, fileHeader := createMultipartFile(t, "testfile.jpg", jpegFileContent)

	// Test storing a file
	t.Run("Store File Successfully", func(t *testing.T) {
		log := loggerPkg.Default
		fileData, err := s3.StoreFile(fileHeader.Filename, file, fileHeader, log)
		assert.NoError(t, err)
		assert.NotNil(t, fileData)
		assert.Contains(t, fileData.Name, fileHeader.Filename)
		defer cleanupS3TestMedia(t, s3, fileData)
	})

	// Test storing an unsupported file type
	t.Run("Unsupported File Type", func(t *testing.T) {
		unsupportedFileContent := []byte("unsupported file content")
		unsupportedFile, unsupportedFileHeader := createMultipartFile(t, "testfile.txt", unsupportedFileContent)

		log := loggerPkg.Default
		_, err := s3.StoreFile(unsupportedFileHeader.Filename, unsupportedFile, unsupportedFileHeader, log)
		assert.Error(t, err)
	})

}

func TestAwsDeleteFile(t *testing.T) {
	s3 := createS3Store(t)

	// Create a test file
	fileData := createS3TestFile(t, s3, "testfile.jpg", jpegFileContent)

	// Test deleting a file
	t.Run("Delete File Successfully", func(t *testing.T) {
		log := loggerPkg.Default
		err := s3.DeleteFile(fileData, log)
		assert.NoError(t, err)
	})
}

func TestAwsListFiles(t *testing.T) {
	s3 := createS3Store(t)

	// Create test files
	fileNames := []string{"file1.jpg", "file2.jpg"}
	createdFiles := make([]fileModels.File, 0)
	for _, fileName := range fileNames {
		fileData := createS3TestFile(t, s3, fileName, jpegFileContent)
		createdFiles = append(createdFiles, *fileData)
	}

	// Test listing files
	t.Run("List Files Successfully", func(t *testing.T) {
		log := loggerPkg.Default
		listed, err := s3.ListFiles(log)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(listed), len(createdFiles))
	})

	// Clean up
	for _, file := range createdFiles {
		cleanupS3TestMedia(t, s3, &file)
	}
}

func TestAwsReadFile(t *testing.T) {
	s3 := createS3Store(t)

	// Create a test file
	fileData := createS3TestFile(t, s3, "testfile.jpg", jpegFileContent)

	// Test reading a file
	t.Run("Read File Successfully", func(t *testing.T) {
		log := loggerPkg.Default

		content, err := s3.ReadFile(fileData, log)
		require.NoError(t, err)
		assert.Equal(t, jpegFileContent, content)
	})

	// Clean up
	defer cleanupS3TestMedia(t, s3, fileData)
}

func TestAwsGetFileInfo(t *testing.T) {
	s3 := createS3Store(t)

	// Create a test file
	fileData := createS3TestFile(t, s3, "testfile.jpg", jpegFileContent)

	// Test getting file info
	t.Run("Get File Info Successfully", func(t *testing.T) {
		log := loggerPkg.Default

		fileInfo, err := s3.GetFileInfo(fileData, log)
		assert.NoError(t, err)
		assert.Equal(t, fileData.Name, fileInfo.Name)
		assert.Equal(t, int64(len(jpegFileContent)), fileInfo.Size)
		assert.Equal(t, "image/jpeg", fileInfo.ContentType)
	})

	// Clean up
	defer cleanupS3TestMedia(t, s3, fileData)
}

// Valid JPEG file content (smallest possible JPEG file)
var jpegFileContent = []byte{
	0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01, 0x01, 0x01, 0x00, 0x48,
	0x00, 0x48, 0x00, 0x00, 0xFF, 0xDB, 0x00, 0x43, 0x00, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xD9,
}
