package testutils

import (
	"io/ioutil"
	"path/filepath"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
)

func GetMockResourceInstance() *MockStruct {
	return &MockStruct{
		SystemData: models.SystemData{
			CreatedByID: uint(9879769658658765678),
			UpdatedByID: uint(999765865865856899),
		},
		Field1: "value1",
		Field2: "value2",
	}
}

func GetTestDB() *database.Database {

	// create a temp db in test.db
	dbPath, err := ioutil.TempDir("", "test-")
	if err != nil {
		panic(err)
	}

	dbPath = filepath.Join(dbPath, "test.db")

	testConfig := &database.DBConfig{
		Driver: "sqlite",
		Path:   dbPath,
		URL:    "not empty",
	}

	db, err := database.LoadDB(testConfig, logger.Default)
	if err != nil {
		panic(err)
	}

	return db
}

func GetTestUser() *models.User {
	return &models.User{
		ID:    uint(999),
		Name:  "Test User",
		Email: "YHs7r@example.com",
		Roles: "admin",
	}
}

func GetTestLogger() *logger.Logger {
	return logger.Default
}
