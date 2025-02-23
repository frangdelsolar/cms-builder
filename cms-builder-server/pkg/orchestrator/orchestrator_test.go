package orchestrator_test

import (
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"

	orc "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/orchestrator"
)

func TestMain(m *testing.M) {
	println("Running pre-test script")

	// load env
	godotenv.Load(".test.env")

	exitCode := m.Run()
	os.Exit(exitCode)
}

func TestInitServer(t *testing.T) {
	o, err := orc.NewOrchestrator()
	assert.NoError(t, err)
	assert.NotNil(t, o.Server)
}

func TestInitDatabase(t *testing.T) {
	o, err := orc.NewOrchestrator()
	assert.NoError(t, err)
	assert.NotNil(t, o.DB)
}

func TestInitConfigReader(t *testing.T) {
	o, err := orc.NewOrchestrator()
	assert.NoError(t, err)
	assert.NotNil(t, o.Config)

	t.Log("Config reader actually reads env variables")
	config := o.Config
	appName := config.GetString(orc.EnvKeys.AppName)
	assert.Equal(t, "test", appName)
}

func TestOrchestratorCreatesLogger(t *testing.T) {
	o, err := orc.NewOrchestrator()
	assert.NoError(t, err)
	assert.NotNil(t, o.Logger)
}
