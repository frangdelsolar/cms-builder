package testing

import (
	"io/ioutil"
	"path/filepath"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	mgr "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"
	"github.com/joho/godotenv"
)

func init() {
	godotenv.Load()
}

type TestUtils struct {
	Db          *database.Database
	Logger      *logger.Logger
	Mgr         *mgr.ResourceManager
	Src         *mgr.Resource
	AdminUser   *models.User
	VisitorUser *models.User
	NoRoleUser  *models.User
}

func NewTestDB() *database.Database {

	// create a temp db in test.db
	dbPath, err := ioutil.TempDir("", "test-")
	if err != nil {
		panic(err)
	}

	dbPath = filepath.Join(dbPath, "test.db")

	// TODO: remove next line
	dbPath = "test.db"

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

func NewTestLogger() *logger.Logger {
	return logger.Default
}
