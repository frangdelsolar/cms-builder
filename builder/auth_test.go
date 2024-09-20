package builder_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/frangdelsolar/cms/builder"
	th "github.com/frangdelsolar/cms/builder/test_helpers"
	"github.com/stretchr/testify/assert"
)

// TestRegisterUserController tests the RegisterUser endpoint by creating a new user and
// verifying the response to make sure the user was created correctly.
func TestRegisterUserController(t *testing.T) {
	t.Log("Testing VerifyUser")
	engine := th.GetDefaultEngine()
	firebase, _ := engine.GetFirebase()
	newUserData := builder.RegisterUserInput{
		Name:     th.RandomName(),
		Email:    th.RandomEmail(),
		Password: th.RandomPassword(),
	}

	bodyBytes, err := json.Marshal(newUserData)
	assert.NoError(t, err)

	header := http.Header{
		"Content-Type": []string{"application/json"},
	}

	responseWriter := th.MockWriter{}
	registerUserRequest := &http.Request{
		Method: http.MethodPost,
		Header: header,
		Body:   io.NopCloser(bytes.NewBuffer(bodyBytes)),
	}

	t.Log("Registering user", newUserData)
	engine.RegisterUserController(&responseWriter, registerUserRequest)

	t.Log("Testing Response")
	userStr := responseWriter.GetWrittenData()
	createdUser := builder.User{}
	json.Unmarshal([]byte(userStr), &createdUser)
	assert.Equal(t, createdUser.Name, newUserData.Name)

	t.Log("Testing Verification token")
	accessToken, err := th.LoginUser(&newUserData)
	assert.NoError(t, err)

	t.Log("Verifying user")
	retrievedUser, err := engine.VerifyUser(accessToken)
	assert.NoError(t, err)
	assert.Equal(t, createdUser.ID, retrievedUser.ID)

	t.Log("Rolling back user registration")
	firebase.RollbackUserRegistration(context.Background(), createdUser.FirebaseId)
}
