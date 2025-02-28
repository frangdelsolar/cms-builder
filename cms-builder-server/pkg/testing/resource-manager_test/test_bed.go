package resourcemanager

import (
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	mgr "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/testing"
)

func SetupHandlerTestBed() TestUtils {
	db := NewTestDB()
	err := db.DB.AutoMigrate(models.User{})
	if err != nil {
		panic(err)
	}

	err = db.DB.AutoMigrate(database.DatabaseLog{})
	if err != nil {
		panic(err)
	}

	log := NewTestLogger()
	mgr := mgr.NewResourceManager(db, log)

	srcConfig := SetupMockResource()
	src, err := mgr.AddResource(srcConfig)
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

	return TestUtils{
		Db:          db,
		Logger:      log,
		Mgr:         mgr,
		Src:         src,
		AdminUser:   admin,
		VisitorUser: visitor,
		NoRoleUser:  noRole,
	}
}
