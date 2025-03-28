package testing

import (
	authModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/models"
	dbModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/models"
	"github.com/joho/godotenv"
)

func SetupDatabaseTestBed() TestUtils {

	godotenv.Load(".test.env")

	db := NewTestDB()
	err := db.DB.AutoMigrate(authModels.User{})
	if err != nil {
		panic(err)
	}

	err = db.DB.AutoMigrate(MockStruct{})
	if err != nil {
		panic(err)
	}

	err = db.DB.AutoMigrate(dbModels.DatabaseLog{})
	if err != nil {
		panic(err)
	}

	log := NewTestLogger()

	return TestUtils{
		Db:        db,
		Logger:    log,
		AdminUser: CreateAdminUser(),
	}
}
