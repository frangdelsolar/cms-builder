package main

import (
	c "cms_server"
)

var log *c.Logger

func main() {
	log = c.GetLogger()
	log.Info().Msg("Starting CMS server")

	server, err := GetServer()
	if err != nil {
		log.Fatal().Err(err).Msg("Error starting server")
	}

	err = server.ListenAndServe()
	if err != nil {
		log.Fatal().Err(err).Msg("Error starting server")
	}
}
