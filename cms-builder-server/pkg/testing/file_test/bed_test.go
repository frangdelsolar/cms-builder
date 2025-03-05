package file_test

import (
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/file"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	mgr "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/store"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/testing"
	"github.com/joho/godotenv"
)

func SetupFileTestBed() TestUtils {

	godotenv.Load(".test.env")

	db := NewTestDB()
	err := db.DB.AutoMigrate(models.User{})
	if err != nil {
		panic(err)
	}

	err = db.DB.AutoMigrate(database.DatabaseLog{})
	if err != nil {
		panic(err)
	}

	err = db.DB.AutoMigrate(models.File{})
	if err != nil {
		panic(err)
	}

	log := NewTestLogger()

	storeConfig := StoreConfig{
		MaxSize:            1024 * 1024 * 1024,
		SupportedMimeTypes: []string{"image/png", "image/jpeg", "image/jpg"},
		MediaFolder:        "test-files",
	}

	localStore, err := NewLocalStore(&storeConfig, "test-files", "http://localhost:8080")
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

	manager := mgr.NewResourceManager(db, log)

	fileSetup := SetupFileResource(manager, db, localStore, log, "http://localhost:8080")
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
