package builder_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/frangdelsolar/cms/builder"
	th "github.com/frangdelsolar/cms/builder/test_helpers"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

const testConfigFilePath = ".test.env"

// TestMain runs the tests, and also performs some pre-test setup (i.e. gets a list of all existing users in
// firebase) and post-test cleanup (i.e. deletes all users that were created during the test).
func TestMain(m *testing.M) {
	println("Running pre-test script")

	// Need to get all the users in firebase
	e, err := th.GetDefaultEngine()

	if err != nil {
		println(err.Error())
		os.Exit(1)
	}

	cli, err := e.Firebase.Auth(context.Background())
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}

	preIterator := cli.Users(context.Background(), "")

	existingUsers := map[string]bool{}
	for {
		user, err := preIterator.Next()
		if err != nil {
			break
		}

		existingUsers[user.UID] = true
	}

	fmt.Printf("Existing users in firebase: %v\n", len(existingUsers))

	// Execute tests
	exitCode := m.Run()

	// Clean up
	println("Running post-test script")
	postIterator := cli.Users(context.Background(), "")

	for {
		user, err := postIterator.Next()
		if err != nil {
			break
		}

		if !existingUsers[user.UID] {
			println(fmt.Sprintf("Deleting user %s", user.UID))
			err = e.Firebase.DeleteUser(context.Background(), user.UID)
			if err != nil {
				println(err.Error())
			}
		}
	}

	os.Exit(exitCode)
}


func TestNewBuilder_ConfigFile(t *testing.T) {

	if os.Getenv("ENVIRONMENT") == "test" || os.Getenv("ENVIRONMENT") == "" {
		godotenv.Load(testConfigFilePath)
	}

	input := &builder.NewBuilderInput{
		ReadConfigFromEnv:    true,
		ReadConfigFromFile:   false,
		ReaderConfigFilePath: testConfigFilePath,
	}

	engine, err := builder.NewBuilder(input)

	assert.NoError(t, err)
	assert.NotNil(t, engine)

	assert.NotNil(t, engine.Admin, "Admin should not be nil")
	assert.NotNil(t, engine.Config, "Config should not be nil")
	assert.NotNil(t, engine.DB, "DB should not be nil")
	assert.NotNil(t, engine.Logger, "Log should not be nil")
	assert.NotNil(t, engine.Server, "Server should not be nil")
	assert.NotNil(t, engine.Firebase, "Firebase should not be nil")
}
