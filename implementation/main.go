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
		ConfigFile: &builder.ConfigFile{
			UseConfigFile: true,
			ConfigPath:    "config.yaml",
		},
	}

	// Build
	engine = builder.NewBuilder(config)

	// Reading a config example
	cfg, err := engine.GetConfigReader()
	if err != nil {
		panic(err)
	}

	loggerConfig := builder.LoggerConfig{
		LogLevel:    cfg.GetString("logLevel"),
		LogFilePath: cfg.GetString("writeToFilePath"),
		WriteToFile: cfg.GetBool("writeToFile"),
	}
	engine.SetLoggerConfig(loggerConfig)

	// Logging example
	log, err = engine.GetLogger()
	if err != nil {
		panic(err)
	}
	log.Info().Msg("Hello World!")

	// DB setup
	dbConfig := builder.DBConfig{
		URL:  "",
		Path: cfg.GetString("dbFile"),
	}
	engine.ConnectDB(&dbConfig)

}
