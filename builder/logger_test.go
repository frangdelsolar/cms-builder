package builder_test

import (
	"os"
	"strings"
	"testing"

	"builder"
)

func TestDefaultLogger(t *testing.T) {

	// Create the logger
	log, err := builder.NewLogger(nil)
	if err != nil {
		t.Error(err)
	}

	// Log a message
	log.Info().Msg("Testing Logger")

	// Read the content of the temporary file
	data, err := os.ReadFile("logs/default.log")
	if err != nil {
		t.Error(err)
	}

	expected := "\"message\":\"Testing Logger\""
	// check if contains
	if !strings.Contains(string(data), expected) {
		t.Errorf("Expected %s, but got %s", expected, string(data))
	}
}

func TestCustomLogger(t *testing.T) {
	// Create the logger
	log, err := builder.NewLogger(&builder.LoggerConfig{
		LogLevel:    "debug",
		WriteToFile: true,
		LogFilePath: "logs/test.log",
	})

	if err != nil {
		t.Error(err)
	}

	// Log a message
	log.Info().Msg("Testing Logger")

	// Read the content of the temporary file
	data, err := os.ReadFile("logs/test.log")
	if err != nil {
		t.Error(err)
	}

	expected := "\"message\":\"Testing Logger\""
	// check if contains
	if !strings.Contains(string(data), expected) {
		t.Errorf("Expected %s, but got %s", expected, string(data))
	}
}
