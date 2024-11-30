package builder_test

import (
	"testing"

	"github.com/frangdelsolar/cms/builder"
	"github.com/stretchr/testify/assert"
)

func TestNewBuilder_SuccessAllOptions(t *testing.T) {
	t.Skip("FIXME")
	input := &builder.NewBuilderInput{
		ReadConfigFromFile: true,
		ConfigFilePath:     "config.yaml", // Replace with a valid config file path
	}

	engine, err := builder.NewBuilder(input)

	assert.NoError(t, err)
	assert.NotNil(t, engine)
	// Additional assertions to verify initialized components (logger, db, server, etc.)
}

func TestNewBuilder_SuccessSomeOptions(t *testing.T) {
	t.Skip("FIXME")

	input := &builder.NewBuilderInput{
		ReadConfigFromFile: true,
		ConfigFilePath:     "config.yaml", // Replace with a valid config file path
	}

	engine, err := builder.NewBuilder(input)

	assert.NoError(t, err)
	assert.NotNil(t, engine)
	// Assert logger is initialized, other components might be nil
}

func TestNewBuilder_MissingConfig(t *testing.T) {
	t.Skip("FIXME")

	var input *builder.NewBuilderInput

	engine, err := builder.NewBuilder(input)

	assert.Error(t, err)
	// assert.EqualError(t, err, builder.ErrBuilderConfigNotProvided.Error())
	assert.Nil(t, engine)
}
