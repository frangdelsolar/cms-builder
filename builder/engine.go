package builder

type EngineServices struct {
	Engine   *Builder
	Admin    *Admin
	DB       *Database
	Server   *Server
	Firebase *FirebaseAdmin
	Log      *Logger
	Config   *ConfigReader
}

func NewDefaultEngine() (EngineServices, error) {
	input := &NewBuilderInput{
		ReadConfigFromFile: true,
		ConfigFilePath:     "config.yaml", // Replace with a valid config file path
		InitializeLogger:   true,
		InitiliazeDB:       true,
		InitiliazeServer:   true,
		InitiliazeAdmin:    true,
		InitiliazeFirebase: true,
		InitiliazeUploader: true,
	}

	var err error
	engine, err := NewBuilder(input)
	if err != nil {
		return EngineServices{}, err
	}

	admin, err := engine.GetAdmin()
	if err != nil {
		return EngineServices{}, err
	}

	db, err := engine.GetDatabase()
	if err != nil {
		return EngineServices{}, err
	}

	server, err := engine.GetServer()
	if err != nil {
		return EngineServices{}, err
	}

	firebase, err := engine.GetFirebase()
	if err != nil {
		return EngineServices{}, err
	}

	logger, err := engine.GetLogger()
	if err != nil {
		return EngineServices{}, err
	}

	configReader, err := engine.GetConfigReader()
	if err != nil {
		return EngineServices{}, err
	}

	return EngineServices{engine, admin, db, server, firebase, logger, configReader}, nil
}
