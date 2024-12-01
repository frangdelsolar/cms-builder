package builder_test

import (
	"os"
	"testing"

	"github.com/frangdelsolar/cms/builder"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

const testConfigFilePath = ".test.env"

func TestNewBuilder_ConfigFile(t *testing.T) {

	if os.Getenv("ENVIRONMENT") == "test" || os.Getenv("ENVIRONMENT") == "" {
		godotenv.Load(testConfigFilePath)
	}

	input := &builder.NewBuilderInput{
		ReadConfigFromEnv:    true,
		ReadConfigFromFile:   false,
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
