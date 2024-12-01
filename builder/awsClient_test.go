package builder_test

import (
	"os"
	"testing"

	"github.com/frangdelsolar/cms/builder"
	th "github.com/frangdelsolar/cms/builder/test_helpers"
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

func TestAWSUploadFile(t *testing.T) {
	e, err := th.GetDefaultEngine()
	assert.NoError(t, err, "GetDefaultEngine should not return an error")

	manager := builder.AwsManager{
		Bucket: e.Config.GetString(builder.EnvKeys.AwsBucket),
	}

	fileName := "test.json"
	file, err := os.ReadFile(testFilePath)
	assert.NoError(t, err, "ReadFile should not return an error")

	err = manager.UploadFile(fileName, file)
	assert.NoError(t, err, "UploadFile should not return an error")
}

func TestAWSDeleteFile(t *testing.T) {
	e, err := th.GetDefaultEngine()
	assert.NoError(t, err, "GetDefaultEngine should not return an error")

	manager := builder.AwsManager{
		Bucket: e.Config.GetString(builder.EnvKeys.AwsBucket),
	}

	fileName := "test-delete.json"

	file, err := os.ReadFile(testFilePath)
	assert.NoError(t, err, "ReadFile should not return an error")

	err = manager.UploadFile(fileName, file)
	assert.NoError(t, err, "UploadFile should not return an error")

	err = manager.DeleteFile(fileName)
	assert.NoError(t, err, "DeleteFile should not return an error")
}
