package testing

import (
	"github.com/joho/godotenv"

	authModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/models"
	dbModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/models"
	fileModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/file/models"
	fileResources "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/file/resources"
	rmPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"
	storePkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/store"
	storeTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/store/types"
)

func SetupFileTestBed() TestUtils {

	godotenv.Load(".test.env")

	db := NewTestDB()
	err := db.DB.AutoMigrate(authModels.User{})
	if err != nil {
		panic(err)
	}

	err = db.DB.AutoMigrate(dbModels.DatabaseLog{})
	if err != nil {
		panic(err)
	}

	err = db.DB.AutoMigrate(fileModels.File{})
	if err != nil {
		panic(err)
	}

	log := NewTestLogger()

	storeConfig := storeTypes.StoreConfig{
		MaxSize:            1024 * 1024 * 1024,
		SupportedMimeTypes: []string{"image/png", "image/jpeg", "image/jpg"},
		MediaFolder:        "test-files",
	}

	localStore, err := storePkg.NewLocalStore(&storeConfig, "test-files", "http://localhost:8080")
	if err != nil {
		panic(err)
	}

	admin := CreateAdminUser()
	visitor := CreateVisitorUser()
	noRole := CreateNoRoleUser()

	err = db.DB.Create(admin).Error
	if err != nil {
		panic(err)
	}

	err = db.DB.Create(visitor).Error
	if err != nil {
		panic(err)
	}

	err = db.DB.Create(noRole).Error
	if err != nil {
		panic(err)
	}

	manager := rmPkg.NewResourceManager(db, log)

	fileSetup := fileResources.SetupFileResource(manager, db, localStore, log, "http://localhost:8080")
	fileResource, err := manager.AddResource(fileSetup)
	if err != nil {
		panic(err)
	}

	return TestUtils{
		Db:          db,
		Logger:      log,
		Store:       localStore,
		AdminUser:   admin,
		VisitorUser: visitor,
		NoRoleUser:  noRole,
		Src:         fileResource,
		Mgr:         manager,
	}
}
