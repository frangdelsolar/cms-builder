package test_helpers

import "github.com/frangdelsolar/cms/builder"

// GetEngineReadyForTests returns a Builder instance with all components initialized.
//
// This function is meant to be used in tests and examples. It will panic if the Builder
// cannot be created.
//
// The returned Builder instance has the following components initialized:
// - Logger
// - Database
// - Server
// - Admin panel
// - Firebase admin
//
// The config file path is set to "config.yaml". You should replace this with a valid
// config file path for your test or example.
func GetEngineReadyForTests() *builder.Builder {
	input := &builder.NewBuilderInput{
		ReadConfigFromFile: true,
		ConfigFilePath:     "config.yaml", // Replace with a valid config file path
		InitializeLogger:   true,
		InitiliazeDB:       true,
		InitiliazeServer:   true,
		InitiliazeAdmin:    true,
		InitiliazeFirebase: true,
	}

	engine, err := builder.NewBuilder(input)

	if err != nil {
		panic(err)
	}

	return engine
}
