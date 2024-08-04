package main

import (
	builder "cms/builder"
)

var config *builder.BuilderConfig
var engine *builder.Builder
var log *builder.Logger

type Example struct {
	Field string
}

func main() {

	config = &builder.BuilderConfig{
		LoggerConfig: &builder.LoggerConfig{
			LogLevel:    "debug",
			LogFilePath: "loggings/defaultee.log",
			WriteToFile: true,
		},
		Environment: "dev",
	}

	engine := builder.NewBuilder(config)

	log = engine.GetLogger()

	// db = engine.GetDatabase()

	log.Info().Msg("Hello World!")
	log.Info().Msgf("Environment: %s", engine.GetEnvironment())

	// log.Info().Msgf("Database: %s", db.AutoMigrate)
	// db.Register(&Example{})

}
