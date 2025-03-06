package testing

import (
	"os"

	"github.com/joho/godotenv"

	authModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/models"
	authResources "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/resources"
	cliPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/clients"
	dbModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/models"
	rmPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"
)

func SetupAuthTestBed() TestUtils {

	if os.Getenv("FIREBASE_SECRET") == "" {
		godotenv.Load(".test.env")
	}

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

	// fb
	fbCfg := &cliPkg.FirebaseConfig{
		Secret: os.Getenv("FIREBASE_SECRET"),
	}
	fb, err := cliPkg.NewFirebaseAdmin(fbCfg)
	if err != nil {
		panic(err)
	}

	getSystemUser := func() *authModels.User {
		return CreateSystemUser()
	}

	srcConfig := authResources.SetupUserResource(fb, db, log, getSystemUser)
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
