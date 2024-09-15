package main

import (
	"github.com/frangdelsolar/cms/builder"

	"gorm.io/gorm"
)

type Example struct {
	*gorm.Model
	Field string `json:"field"`
}

func main() {
	builderCfg := builder.NewBuilderInput{
		ReadConfigFromFile: true,
		ConfigFilePath:     "config.yaml",
		InitializeLogger:   true,
		InitiliazeDB:       true,
		InitiliazeServer:   true,
		InitiliazeAdmin:    true,
		InitiliazeFirebase: true,
	}
	engine, err := builder.NewBuilder(&builderCfg)
	if err != nil {
		panic(err)
	}

	log, err := engine.GetLogger()
	if err != nil {
		panic(err)
	}
	log.Info().Msg("Logger initialized correctly")

	db, err := engine.GetDatabase()
	if err != nil {
		panic(err)
	}
	log.Info().Interface("Database", db).Msg("Database initialized correctly")

	server, err := engine.GetServer()
	if err != nil {
		panic(err)
	}

	log.Info().Msg("Server initialized correctly")

	admin, err := engine.GetAdmin()
	if err != nil {
		panic(err)
	}
	admin.Register(&Example{})

	fb, err := engine.GetFirebase()
	if err != nil {
		panic(err)
	}
	log.Info().Interface("Firebase", fb).Msg("Firebase initialized correctly")

	server.Run()

}
