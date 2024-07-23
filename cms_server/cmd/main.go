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

	cfg := cms.Config{
		Logger: log.Logger,
		DB:     db.DB,
	}

	cms.Setup(&cfg)

	cms.Register(&Primitive{})

	server, err := GetServer()
	if err != nil {
		log.Fatal().Err(err).Msg("Error starting server")
	}

	err = server.ListenAndServe()
	if err != nil {
		log.Fatal().Err(err).Msg("Error starting server")
	}

}
