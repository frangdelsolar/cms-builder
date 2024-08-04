package builder

type Builder struct {
	env    string
	logger *Logger
	// db     *Database

	// server *Server
	// router *Router
}

type BuilderConfig struct {
	Environment string
	*LoggerConfig
	// *DatabaseConfig
}

func NewBuilder(cfg *BuilderConfig) *Builder {

	// Environment
	if cfg.Environment == "" {
		cfg.Environment = "dev"
	}

	// Logger
	logger := NewLogger(cfg.LoggerConfig)

	// if cfg.DatabaseConfig == nil {
	// 	cfg.DatabaseConfig = &DatabaseConfig{}
	// }
	// dbConfig := cfg.DatabaseConfig
	// dbConfig.AppEnv = cfg.Environment

	// db, err := NewDB(dbConfig)
	// if err != nil {
	// 	logger.Fatal().Err(err).Msg("Failed to initialize database")
	// }

	return &Builder{
		logger: logger,
		env:    cfg.Environment,
		// db:     db,
	}
}

func (b *Builder) GetLogger() *Logger {
	return b.logger
}

// func (b *Builder) GetDatabase() *Database {
// 	return b.db
// }

func (b *Builder) GetEnvironment() string {
	return b.env
}
