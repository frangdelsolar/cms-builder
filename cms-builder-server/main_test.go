package orchestrator_test

import (
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func TestMain(m *testing.M) {
	println("Running pre-test script")

	// load env
	godotenv.Load(".test.env")

	exitCode := m.Run()
	os.Exit(exitCode)
}
