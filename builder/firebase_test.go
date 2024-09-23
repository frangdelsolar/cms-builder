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
	e := th.GetDefaultEngine()
	firebase, err := e.Engine.GetFirebase()

	assert.NoError(t, err)
	assert.NotNil(t, firebase)
}

// TestRegisterFirebaseUser tests the registration of a new user in Firebase, and
// the rolling back of the user registration.
func TestRegisterFirebaseUser(t *testing.T) {
	t.Log("Testing Firebase User registration and rollback")
	e := th.GetDefaultEngine()
	newUserData := builder.RegisterUserInput{
		Name:     th.RandomName(),
		Email:    th.RandomEmail(),
		Password: th.RandomPassword(),
	}

	t.Log("Registering user", newUserData)
	user, err := e.Firebase.RegisterUser(context.Background(), newUserData)
	assert.NoError(t, err)
	assert.NotNil(t, user)

	// Perform rollback
	t.Log("Rolling back user registration", user.UID)
	err = e.Firebase.RollbackUserRegistration(context.Background(), user.UID)
	assert.NoError(t, err)
}

func TestLoginUser(t *testing.T) {
	t.Log("Testing Firebase User login")
	e := th.GetDefaultEngine()

	newUserData := builder.RegisterUserInput{
		Name:     th.RandomName(),
		Email:    th.RandomEmail(),
		Password: th.RandomPassword(),
	}

	t.Log("Registering user", newUserData)
	fbUser, err := e.Firebase.RegisterUser(context.Background(), newUserData)
	assert.NoError(t, err)

	t.Log("Logging in user", fbUser.UID)
	token, err := th.LoginUser(&newUserData)
	assert.NoError(t, err)

	t.Log("Testing Verification token")
	tkn, err := e.Firebase.VerifyIDToken(context.Background(), token)
	assert.NoError(t, err)
	assert.Equal(t, fbUser.UID, tkn.UID)

	// Perform rollback
	t.Log("Rolling back user registration", fbUser.UID)
	err = e.Firebase.RollbackUserRegistration(context.Background(), fbUser.UID)
	assert.NoError(t, err)
}
