package testing

import (
	"os"

	"github.com/joho/godotenv"

	authModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/models"
	cliPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/clients"
	dbModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/models"
	rlModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/request-logger/models"
	rmPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"
	svrPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
	svrTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server/types"
)

func SetupServerTestBed() TestUtils {

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

	err = db.DB.AutoMigrate(rlModels.RequestLog{})
	if err != nil {
		panic(err)
	}

	err = db.DB.AutoMigrate(MockStruct{})
	if err != nil {
		panic(err)
	}

	log := NewTestLogger()

	firebaseConfig := &cliPkg.FirebaseConfig{
		Secret: os.Getenv("FIREBASE_SECRET"),
	}
	firebaseClient, err := cliPkg.NewFirebaseAdmin(firebaseConfig)
	if err != nil {
		panic(err)
	}

	serverConfig := svrTypes.ServerConfig{
		Host:      os.Getenv("SERVER_HOST"),
		Port:      os.Getenv("SERVER_PORT"),
		CsrfToken: os.Getenv("CSRF_TOKEN"),
		AllowedOrigins: []string{
			"*",
		},
		GodToken: os.Getenv("GOD_TOKEN"),
		GodUser:  CreateGodUser(),
		SystemUser: &authModels.User{
			Email: "system",
		},
		Firebase: firebaseClient,
	}

	server, err := svrPkg.NewServer(&serverConfig, db, log)
	if err != nil {
		panic(err)
	}

	mgr := rmPkg.NewResourceManager(db, log)

	srcConfig := SetupMockResource()
	_, err = mgr.AddResource(srcConfig)
	if err != nil {
		panic(err)
	}

	return TestUtils{
		Db:          db,
		Logger:      log,
		Server:      server,
		Mgr:         mgr,
		AdminUser:   CreateAdminUser(),
		VisitorUser: CreateVisitorUser(),
	}
}
