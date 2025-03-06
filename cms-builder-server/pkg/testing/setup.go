package testing

import (
	"github.com/joho/godotenv"

	authModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/models"
	dbPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"
	loggerPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger"
	loggerTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger/types"
	rmPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"
	rmTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager/types"
	scheduler "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/scheduler"
	serverTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server/types"
	storeTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/store/types"
)

func init() {
	godotenv.Load(".test.env")
}

type TestUtils struct {
	Db            *dbTypes.DatabaseConnection
	Logger        *loggerTypes.Logger
	Mgr           *rmPkg.ResourceManager
	Src           *rmTypes.Resource
	AdminUser     *authModels.User
	VisitorUser   *authModels.User
	NoRoleUser    *authModels.User
	Scheduler     *scheduler.Scheduler
	SchedulerUser *authModels.User
	Server        *serverTypes.Server
	Store         storeTypes.Store
}

func NewTestDB() *dbTypes.DatabaseConnection {

	dbPath := "test.db"

	testConfig := &dbTypes.DatabaseConfig{
		Driver: "sqlite",
		Path:   dbPath,
		URL:    "not empty",
	}

	db, err := dbPkg.NewDatabaseConnection(testConfig, loggerPkg.Default)
	if err != nil {
		panic(err)
	}

	return db
}

func NewTestLogger() *loggerTypes.Logger {
	return loggerPkg.Default
}
