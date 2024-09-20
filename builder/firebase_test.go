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
	engine := th.GetDefaultEngine()
	firebase, err := engine.GetFirebase()

	assert.NoError(t, err)
	assert.NotNil(t, firebase)
}

func TestRegisterFirebaseUser(t *testing.T) {
	t.Log("Testing Firebase User registration and rollback")
	engine := th.GetDefaultEngine()
	firebase, _ := engine.GetFirebase()
	newUserData := builder.RegisterUserInput{
		Name:     th.RandomName(),
		Email:    th.RandomEmail(),
		Password: th.RandomPassword(),
	}

	t.Log("Registering user", newUserData)
	user, err := firebase.RegisterUser(context.Background(), newUserData)
	assert.NoError(t, err)
	assert.NotNil(t, user)

	// Perform rollback
	t.Log("Rolling back user registration", user.UID)
	err = firebase.RollbackUserRegistration(context.Background(), user.UID)
	assert.NoError(t, err)
}

func TestLoginUser(t *testing.T) {
	t.Log("Testing Firebase User login")
	engine := th.GetDefaultEngine()
	firebase, _ := engine.GetFirebase()
	newUserData := builder.RegisterUserInput{
		Name:     th.RandomName(),
		Email:    th.RandomEmail(),
		Password: th.RandomPassword(),
	}

	t.Log("Registering user", newUserData)
	fbUser, err := firebase.RegisterUser(context.Background(), newUserData)
	assert.NoError(t, err)

	t.Log("Logging in user", fbUser.UID)
	token, err := th.LoginUser(&newUserData)
	assert.NoError(t, err)

	t.Log("Testing Verification token")
	tkn, err := firebase.VerifyIDToken(context.Background(), token)
	assert.NoError(t, err)
	assert.Equal(t, fbUser.UID, tkn.UID)

	// Perform rollback
	t.Log("Rolling back user registration", fbUser.UID)
	err = firebase.RollbackUserRegistration(context.Background(), fbUser.UID)
	assert.NoError(t, err)
}
