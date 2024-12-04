package builder_test

import (
	"testing"

	"github.com/frangdelsolar/cms/builder"
	th "github.com/frangdelsolar/cms/builder/test_helpers"

	"github.com/stretchr/testify/assert"
)

func TestGetPostmanEnv(t *testing.T) {
	e, err := th.GetDefaultEngine()
	assert.NoError(t, err, "GetDefaultEngine should not return an error")

	postmanEnv, err := e.Engine.GetPostmanEnv()
	assert.NoError(t, err, "GetPostmanEnv should not return an error")

	assert.NotNil(t, postmanEnv, "GetPostmanEnv should not return nil")
	assert.Equal(t, postmanEnv.Name, "test", "PostmanEnv name should be 'test'")
}

func TestGetPostmanCollection(t *testing.T) {
	e, err := th.GetDefaultEngine()
	assert.NoError(t, err, "GetDefaultEngine should not return an error")

	collection, err := e.Engine.GetPostmanCollection()
	assert.NoError(t, err, "GetPostmanCollection should not return an error")

	assert.NotNil(t, collection, "GetPostmanCollection should not return nil")
	assert.Equal(t, collection.Info.Name, "test", "PostmanEnv name should be 'test'")
}

func TestExportPostman(t *testing.T) {
	e, err := th.GetDefaultEngine()
	assert.NoError(t, err, "GetDefaultEngine should not return an error")

	err = e.Engine.ExportPostman()
	assert.NoError(t, err, "ExportPostman should not return an error")

	assert.FileExists(t, builder.PostmanSchemaFilePath, "PostmanSchemaFilePath should exist")
	assert.FileExists(t, builder.PostmanEnvFilePath, "PostmanEnvFilePath should exist")
}
