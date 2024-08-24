package main

import (
	builder "cms/builder"

	"gorm.io/gorm"
)

type Example struct {
	*gorm.Model
	Field string `json:"field"`
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
		LogFilePath: cfg.GetString("logFilePath"),
		WriteToFile: cfg.GetBool("logWriteToFile"),
	}
	engine.SetLoggerConfig(loggerConfig)

	// Logging example
	log, err = engine.GetLogger()
	if err != nil {
		panic(err)
	}
	log.Info().Msg("Logger setup")

	// DB setup
	dbConfig := builder.DBConfig{
		URL:  "",
		Path: cfg.GetString("dbFile"),
	}
	engine.ConnectDB(&dbConfig)

	log.Debug().Interface("DBConfig", dbConfig).Msg("DB setup")

	// Server setup
	serverConfig := builder.ServerConfig{
		Host: cfg.GetString("host"),
		Port: cfg.GetString("port"),
	}
	err = engine.SetServerConfig(serverConfig)
	if err != nil {
		log.Error().Err(err).Msg("Error setting up server")
		panic(err)
	}

	svr, err := engine.GetServer()
	if err != nil {
		log.Error().Err(err).Msg("Error getting server")
		panic(err)
	}

	// Admin setup
	err = engine.SetupAdmin()
	if err != nil {
		log.Error().Err(err).Msg("Error setting up admin panel")
		panic(err)
	}

	admin := engine.GetAdmin()
	admin.Register(&Example{})

	engine.SetupFirebase()

	svr.Run()

}
