package builder

var log *Logger

type Builder struct {
	env    string
	logger *Logger
}

type BuilderConfig struct {
	Environment string
	*LoggerConfig
}

func NewBuilder(cfg *BuilderConfig) *Builder {

	var output = &Builder{}

	// Logger
	log = NewLogger(cfg.LoggerConfig)
	output.logger = log

	// Load config
	err := output.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Error loading config")
	}

	// Environment
	if cfg.Environment == "" {
		cfg.Environment = "dev"
	}
	output.env = cfg.Environment

	return output
}

func (b *Builder) GetLogger() *Logger {
	return b.logger
}

func (b *Builder) GetEnvironment() string {
	return b.env
}
