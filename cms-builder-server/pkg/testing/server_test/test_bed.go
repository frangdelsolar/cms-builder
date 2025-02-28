package server

import (
	"os"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/clients"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/testing"
	"github.com/joho/godotenv"
)

func SetupServerTestBed() TestUtils {

	godotenv.Load(".test.env")

	db := NewTestDB()
	err := db.DB.AutoMigrate(models.User{})
	if err != nil {
		panic(err)
	}

	err = db.DB.AutoMigrate(models.DatabaseLog{})
	if err != nil {
		panic(err)
	}

	log := NewTestLogger()

	firebaseConfig := &clients.FirebaseConfig{
		Secret: os.Getenv("FIREBASE_SECRET"),
	}
	firebaseClient, err := clients.NewFirebaseAdmin(firebaseConfig)
	if err != nil {
		panic(err)
	}

	serverConfig := ServerConfig{
		Host:      "localhost",
		Port:      "8080",
		CsrfToken: "secret",
		AllowedOrigins: []string{
			"*",
		},
		GodToken: "secret",
		GodUser:  CreateGodUser(),
		SystemUser: &models.User{
			Email: "system",
		},
		Firebase: firebaseClient,
	}

	server, err := NewServer(&serverConfig, db, log)
	if err != nil {
		panic(err)
	}

	return TestUtils{
		Db:     db,
		Logger: log,
		Server: server,
	}
}
