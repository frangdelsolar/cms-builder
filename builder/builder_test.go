package builder_test

import (
	"testing"

	"github.com/frangdelsolar/cms/builder"
	"github.com/stretchr/testify/assert"
)

const testConfigFilePath = ".test.env"

func TestNewBuilder_ConfigFile(t *testing.T) {
	input := &builder.NewBuilderInput{
		ReadConfigFromFile:   true,
		ReaderConfigFilePath: testConfigFilePath,
	}

	engine, err := builder.NewBuilder(input)

	assert.NoError(t, err)
	assert.NotNil(t, engine)

	assert.NotNil(t, engine.Admin, "Admin should not be nil")
	assert.NotNil(t, engine.Config, "Config should not be nil")
	assert.NotNil(t, engine.DB, "DB should not be nil")
	assert.NotNil(t, engine.Logger, "Log should not be nil")
	assert.NotNil(t, engine.Server, "Server should not be nil")
	assert.NotNil(t, engine.Firebase, "Firebase should not be nil")
}
