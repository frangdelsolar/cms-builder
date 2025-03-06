package testing

import (
	authModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/models"
	dbModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/models"
	rmPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"
)

func SetupHandlerTestBed() TestUtils {
	db := NewTestDB()
	err := db.DB.AutoMigrate(authModels.User{})
	if err != nil {
		panic(err)
	}

	err = db.DB.AutoMigrate(dbModels.DatabaseLog{})
	if err != nil {
		panic(err)
	}

	log := NewTestLogger()
	mgr := rmPkg.NewResourceManager(db, log)

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

	err = db.DB.Create(noRole).Error
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
