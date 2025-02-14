package builder_test

import (
	"os"
	"testing"

	builder "github.com/frangdelsolar/cms-builder/cms-builder-server"
	th "github.com/frangdelsolar/cms-builder/cms-builder-server/test_helpers"
	"github.com/stretchr/testify/assert"
)

func TestAWSIsReady(t *testing.T) {
	e, err := th.GetDefaultEngine()
	assert.NoError(t, err, "GetDefaultEngine should not return an error")

	manager := builder.AwsManager{
		Bucket: e.Config.GetString(builder.EnvKeys.AwsBucket),
	}

	ready := manager.IsReady()
	assert.True(t, ready, "AWS should be ready")
}

func TestAWSUploadDeleteFile(t *testing.T) {
	e, err := th.GetDefaultEngine()
	assert.NoError(t, err, "GetDefaultEngine should not return an error")

	manager := builder.AwsManager{
		Bucket: e.Config.GetString(builder.EnvKeys.AwsBucket),
	}

	fileName := "test-upload.json"
	directory := "test_output"

	file, err := os.ReadFile(testFilePath)
	assert.NoError(t, err, "ReadFile should not return an error")

	t.Log("Testing S3 file upload")
	err = manager.UploadFile(directory, fileName, file)
	assert.NoError(t, err, "UploadFile should not return an error")

	path := directory + "/" + fileName
	data, err := manager.DownloadFile(path)
	assert.NoError(t, err, "DownloadFile should not return an error")
	assert.NotNil(t, data, "DownloadFile should not return nil")

	t.Log("Testing S3 file delete")
	err = manager.DeleteFile(fileName)
	assert.NoError(t, err, "DeleteFile should not return an error")
}
