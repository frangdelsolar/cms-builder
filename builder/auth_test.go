package builder_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/frangdelsolar/cms/builder"
	th "github.com/frangdelsolar/cms/builder/test_helpers"
	"github.com/stretchr/testify/assert"
)

// TestRegisterUserController tests the RegisterUser endpoint by creating a new user and
// verifying the response to make sure the user was created correctly.
func TestRegisterUserController(t *testing.T) {
	t.Log("Testing VerifyUser")
	e, err := th.GetDefaultEngine()
	assert.NoError(t, err, "GetDefaultEngine should not return an error")

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

	t.Log("Registering user")
	e.Engine.RegisterUserController(&responseWriter, registerUserRequest)

	t.Log("Testing Response")
	createdUser := builder.User{}
	response, err := builder.ParseResponse(responseWriter.Buffer.Bytes(), &createdUser)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, createdUser.Name, newUserData.Name)

	t.Log("Testing Verification token")
	accessToken, err := th.LoginUser(&newUserData)
	assert.NoError(t, err)

	t.Log("Verifying user")
	retrievedUser, err := e.Engine.VerifyUser(accessToken)
	assert.NoError(t, err)
	assert.Equal(t, createdUser.ID, retrievedUser.ID)

	t.Log("Rolling back user registration")
	e.Firebase.RollbackUserRegistration(context.Background(), createdUser.FirebaseId)
}

// TestAuthenticationMiddleware tests the authentication middleware by registering a user,
// logging in with that user, and verifying that the middleware adds the "auth" header
// to the request.
func TestAuthenticationMiddleware(t *testing.T) {
	e, err := th.GetDefaultEngine()
	assert.NoError(t, err, "GetDefaultEngine should not return an error")

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("auth")
		if authHeader != "true" {
			t.Errorf("missing auth header")
		}
	})

	handlerToTest := e.Engine.AuthMiddleware(nextHandler)

	req := httptest.NewRequest("GET", "http://testing", nil)

	userData := th.RandomUserData()
	_, rollback := th.RegisterTestUser(userData)
	defer rollback()

	accessToken, err := th.LoginUser(userData)
	assert.NoError(t, err, "Error logging in user")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	handlerToTest.ServeHTTP(httptest.NewRecorder(), req)
}
