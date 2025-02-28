package testing

import (
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	mgr "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/scheduler"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
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

	Scheduler     *scheduler.Scheduler
	SchedulerUser *models.User

	Server *server.Server
}

func NewTestDB() *database.Database {

	dbPath := "test.db"

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
