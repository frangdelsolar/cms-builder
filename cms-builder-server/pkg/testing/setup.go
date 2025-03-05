package testing

import (
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger"
	loggerTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger/types"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	mgr "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/scheduler"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/store"
	"github.com/joho/godotenv"
)

func init() {
	godotenv.Load()
}

type TestUtils struct {
	Db          *dbTypes.DatabaseConnection
	Logger      *loggerTypes.Logger
	Mgr         *mgr.ResourceManager
	Src         *mgr.Resource
	AdminUser   *models.User
	VisitorUser *models.User
	NoRoleUser  *models.User

	Scheduler     *scheduler.Scheduler
	SchedulerUser *models.User

	Server *server.Server

	Store store.Store
}

func NewTestDB() *dbTypes.DatabaseConnection {

	dbPath := "test.db"

	testConfig := &dbTypes.DatabaseConfig{
		Driver: "sqlite",
		Path:   dbPath,
		URL:    "not empty",
	}

	db, err := database.NewDatabaseConnection(testConfig, logger.Default)
	if err != nil {
		panic(err)
	}

	return db
}

func NewTestLogger() *loggerTypes.Logger {
	return logger.Default
}
