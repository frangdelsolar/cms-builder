package builder_test

import (
	"context"
	"testing"

	builder "github.com/frangdelsolar/cms/cms-builder-server"
	th "github.com/frangdelsolar/cms/cms-builder-server/test_helpers"
	"github.com/stretchr/testify/assert"
)

func TestNewFirebaseAdmin_Success(t *testing.T) {

	t.Log("Testing Firebase Admin initialization")
	e, err := th.GetDefaultEngine()
	assert.NoError(t, err, "GetDefaultEngine should not return an error")
	firebase := e.Firebase
	assert.NotNil(t, firebase)
}

// TestRegisterFirebaseUser tests the registration of a new user in Firebase, and
// the rolling back of the user registration.
func TestRegisterFirebaseUser(t *testing.T) {
	t.Log("Testing Firebase User registration and rollback")
	e, err := th.GetDefaultEngine()
	assert.NoError(t, err, "GetDefaultEngine should not return an error")
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
	e, err := th.GetDefaultEngine()
	assert.NoError(t, err, "GetDefaultEngine should not return an error")

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
