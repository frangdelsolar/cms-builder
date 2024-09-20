package test_helpers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/frangdelsolar/cms/builder"
)

// NewRequest creates a new HTTP request with the given method and body.
// If authenticate is true, the function will create a new user, log in the user,
// and add the authentication token and the user ID to the request headers.
// The function returns a pointer to the created request and a function that
// can be used to undo the user registration.
func NewRequest(method string, body string, authenticate bool, user *builder.User) (*http.Request, *builder.User, func()) {
	callback := func() {}

	header := http.Header{
		"Content-Type": []string{"application/json"},
	}
	if authenticate {
		userData := RandomUserData()
		user, rollback := RegisterTestUser(userData)
		token, err := LoginUser(userData)
		if err != nil {
			fmt.Printf("Error: %+v\n", err)
			panic(err)
		}
		header.Add("Authorization", "Bearer "+token)
		header.Set("requested_by", fmt.Sprint(user.ID))
		callback = rollback
	}
	return &http.Request{
		Method: method,
		Header: header,
		Body:   io.NopCloser(bytes.NewBuffer([]byte(body))),
	}, user, callback
}

// RegisterTestUser registers a new user in the default engine, and returns a pointer to the created user
// and a function that can be used to undo the user registration.
//
// Parameters:
// - newUserData: The RegisterUserInput containing the data to register the user with.
//
// Returns:
// - *builder.User: The created user.
// - func(): A function that can be used to undo the user registration.
func RegisterTestUser(newUserData *builder.RegisterUserInput) (*builder.User, func()) {
	engine := GetDefaultEngine()
	firebase, _ := engine.GetFirebase()
	bodyBytes, _ := json.Marshal(newUserData)

	header := http.Header{
		"Content-Type": []string{"application/json"},
	}

	responseWriter := MockWriter{}
	registerUserRequest := &http.Request{
		Method: http.MethodPost,
		Header: header,
		Body:   io.NopCloser(bytes.NewBuffer(bodyBytes)),
	}

	engine.RegisterUserController(&responseWriter, registerUserRequest)

	userStr := responseWriter.GetWrittenData()
	createdUser := builder.User{}
	json.Unmarshal([]byte(userStr), &createdUser)

	return &createdUser, func() {
		firebase.RollbackUserRegistration(context.Background(), createdUser.FirebaseId)
	}
}
