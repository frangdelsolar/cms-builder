package orchestrator_test

import (
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"

	orc "github.com/frangdelsolar/cms-builder/cms-builder-server"
	authModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/models"
	dbModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/models"
	rlModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/request-logger/models"
)

func TestMain(m *testing.M) {
	println("Running pre-test script")

	// load env
	godotenv.Load(".test.env")

	exitCode := m.Run()
	os.Exit(exitCode)
}

func TestNewOrchestrator(t *testing.T) {
	o, err := orc.NewOrchestrator()
	assert.NoError(t, err)
	assert.NotNil(t, o.Config)

	// Config Reader
	config := o.Config
	appName := config.GetString(orc.EnvKeys.AppName)
	assert.Equal(t, "test", appName)

	// Logger
	log := o.Logger
	assert.NotNil(t, log)
	log.Info().Msg("test log")

	// Database
	db := o.DB
	assert.NotNil(t, db)

	// Firebase
	firebase := o.FirebaseClient
	assert.NotNil(t, firebase)

	// ResourceManager
	resourceManager := o.ResourceManager
	assert.NotNil(t, resourceManager)

	// Auth
	userResource, err := resourceManager.GetResource(authModels.User{})
	assert.NoError(t, err)
	assert.NotNil(t, userResource)

	// Database Logger
	dbResource, err := resourceManager.GetResource(dbModels.DatabaseLog{})
	assert.NoError(t, err)
	assert.NotNil(t, dbResource)

	// Request Logger
	reqResource, err := resourceManager.GetResource(rlModels.RequestLog{})
	assert.NoError(t, err)
	assert.NotNil(t, reqResource)

	// Users
	users := o.Users
	assert.NotNil(t, users.God)
	assert.NotNil(t, users.Scheduler)
	assert.NotNil(t, users.System)
	assert.NotNil(t, users.Admin)

	// Server
	server := o.Server
	assert.NotNil(t, server)

	// Store
	store := o.Store
	assert.NotNil(t, store)

	// Scheduler
	scheduler := o.Scheduler
	assert.NotNil(t, scheduler)

}
