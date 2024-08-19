package main

import (
	builder "cms/builder"
)

type Example struct {
	Field string
}

func main() {
	var config *builder.BuilderConfig
	var engine *builder.Builder
	var log *builder.Logger

	// Setup
	config = &builder.BuilderConfig{
		LoggerConfig: &builder.LoggerConfig{
			LogLevel:    "debug",
			LogFilePath: "logs/default.log",
			WriteToFile: true,
		},
		ConfigFile: &builder.ConfigFile{
			UseConfigFile: true,
			ConfigPath:    "config.yaml",
		},
	}

	// Build
	engine = builder.NewBuilder(config)

	// Logging example
	log, err := engine.GetLogger()
	if err != nil {
		panic(err)
	}
	log.Info().Msg("Hello World!")

	// Reading a config example
	cfg, err := engine.GetConfigReader()
	if err != nil {
		panic(err)
	}

	// DB setup
	dbConfig := builder.DBConfig{
		URL:  "",
		Path: cfg.GetString("dbFile"),
	}
	engine.ConnectDB(&dbConfig)

}
