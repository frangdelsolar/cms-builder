package builder_test

import (
	"context"
	"testing"

	"github.com/frangdelsolar/cms/builder"
	th "github.com/frangdelsolar/cms/builder/test_helpers"
	"github.com/stretchr/testify/assert"
)

func TestNewFirebaseAdmin_Success(t *testing.T) {
	t.Log("Testing Firebase Admin initialization")
	engine := th.GetEngineReadyForTests()

	admin, err := engine.GetFirebase()

	assert.NoError(t, err)
	assert.NotNil(t, admin)
}

func TestRegisterUser(t *testing.T) {
	t.Log("Testing Firebase User registration and rollback")
	engine := th.GetEngineReadyForTests()

	admin, _ := engine.GetFirebase()

	newUserData := builder.RegisterUserInput{
		Name:     th.RandomName(),
		Email:    th.RandomEmail(),
		Password: th.RandomPassword(),
	}
	t.Log("Registering user", newUserData)

	user, err := admin.RegisterUser(context.Background(), newUserData)
	assert.NoError(t, err)
	assert.NotNil(t, user)

	// Perform rollback
	t.Log("Rolling back user registration", user.UID)

	err = admin.RollbackUserRegistration(context.Background(), user.UID)
	assert.NoError(t, err)
}

func TestLoginUser(t *testing.T) {
	t.Log("Testing Firebase User login")

	engine := th.GetEngineReadyForTests()
	admin, _ := engine.GetFirebase()
	newUserData := builder.RegisterUserInput{
		Name:     th.RandomName(),
		Email:    th.RandomEmail(),
		Password: th.RandomPassword(),
	}

	fbUser, err := admin.RegisterUser(context.Background(), newUserData)
	assert.NoError(t, err)

	err = th.LoginUser(&newUserData)
	t.Log("Logging in user", fbUser.UID)
	assert.NoError(t, err)

	// Perform rollback
	t.Log("Rolling back user registration", fbUser.UID)
	err = admin.RollbackUserRegistration(context.Background(), fbUser.UID)
	assert.NoError(t, err)
}
