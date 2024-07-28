package main

import (
	cms "cms_server/cms_admin"
)

func main() {

	config, err := GetConfig()
	if err != nil {
		panic(err)
	}

	log = GetLogger()
	log.Info().Msg("Starting Test to CMS")

	db, err := LoadDB()
	if err != nil {
		log.Fatal().Err(err).Msg("Error loading database")
	}

	server, err := GetServer()
	if err != nil {
		log.Fatal().Err(err).Msg("Error starting server")
	}

	// Setup cms
	cfg := cms.Config{
		Logger:  log.Logger,
		DB:      db.DB,
		RootDir: config.RootDir,
	}
	err = cms.Setup(&cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Error setting up cms")
		return
	}
	// Register models into cms
	cms.Register(&Primary{})

	// Append cms routes to server
	cms.Routes(server.Router())

	err = server.Run()
	if err != nil {
		log.Fatal().Err(err).Msg("Error starting server")
	}

}
