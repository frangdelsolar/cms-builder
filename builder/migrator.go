package builder

import (
	"embed"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/lib/pq"
)

//go:embed migrations/*
var migrations embed.FS

func Migrate(dbConfig *DBConfig) error {

	dbUrl := dbConfig.URL

	if dbConfig.Driver == "sqlite" {
		dbUrl = "sqlite://" + dbConfig.Path
	}

	log.Info().Str("url", dbUrl).Msg("Migrating database...")

	// Use iofs to create a migration source from the embedded files.
	driver, err := iofs.New(migrations, "migrations")
	if err != nil {
		log.Error().Err(err).Msg("Error creating iofs migration source")
		return fmt.Errorf("creating iofs migration source: %w", err)
	}

	m, err := migrate.NewWithSourceInstance(
		"iofs",
		driver,
		dbUrl,
	)

	if err != nil {
		log.Error().Err(err).Msg("Error creating migrate instance")
		return err
	}

	err = m.Up()
	if err != nil {

		if err == migrate.ErrNoChange {
			log.Info().Msg("Database already up to date")
			return nil
		}

		log.Error().Err(err).Msg("Error migrating database")
		return err
	}
	// m.Steps(2)

	return nil
}
