package auth_test

import (
	"os"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/clients"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	mgr "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/testing"
	"github.com/joho/godotenv"
)

func SetupAuthTestBed() TestUtils {

	if os.Getenv("FIREBASE_SECRET") == "" {
		godotenv.Load(".test.env")
	}

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

	// fb
	fbCfg := &clients.FirebaseConfig{
		Secret: os.Getenv("FIREBASE_SECRET"),
	}
	fb, err := clients.NewFirebaseAdmin(fbCfg)
	if err != nil {
		panic(err)
	}

	getSystemUser := func() *models.User {
		return CreateSystemUser()
	}

	srcConfig := auth.SetupUserResource(fb, db, log, getSystemUser)
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
