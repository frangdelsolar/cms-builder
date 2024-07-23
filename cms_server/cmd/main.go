package main

import (
	cms "cms_server"
)

func main() {
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
		Logger: log.Logger,
		DB:     db.DB,
	}
	cms.Setup(&cfg)
	// Register models into cms
	cms.Register(&Primitive{})

	// Append cms routes to server
	cms.Routes(server.Router())

	err = server.ListenAndServe()
	if err != nil {
		log.Fatal().Err(err).Msg("Error starting server")
	}

}
