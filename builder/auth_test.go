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
	e := th.GetDefaultEngine()

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

// func TestAuthenticationMiddleware(t *testing.T) {
// 	// Create a mock HTTP server
// 	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		// Your handler code here
// 	}))
// 	defer ts.Close()

// 	// Create a middleware instance
// 	authMiddleware := middleware.NewAuthenticationMiddleware()

// 	// Create a handler function
// 	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		// Your handler code here
// 	})

// 	// Wrap the handler with the middleware
// 	wrappedHandler := authMiddleware.Handle(handler)

// 	// Send a request to the wrapped handler
// 	resp, err := http.Get(ts.URL)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	defer resp.Body.Close()

// 	// Assert the response
// 	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
// }
