package test_helpers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/frangdelsolar/cms/builder"
	"github.com/gorilla/mux"
)

// NewRequest creates a new HTTP request with the given method and body.
// If authenticate is true, the function will create a new user, log in the user,
// and add the authentication token and the user ID to the request headers.
// The function returns a pointer to the created request and a function that
// can be used to undo the user registration.
func NewRequest(method string, body string, authenticate bool, user *builder.User, vars map[string]string) (*http.Request, *builder.User, func()) {
	callback := func() {}

	header := http.Header{
		"Content-Type": []string{"application/json"},
	}

	if user == nil && authenticate {
		userData := RandomUserData()
		newUser, rollback := RegisterTestUser(userData)
		callback = rollback
		user = newUser
	}

	if authenticate {
		accessToken, err := LoginUser(&builder.RegisterUserInput{
			Email:    user.Email,
			Password: "password123",
		})
		if err != nil {
			panic(err)
		}
		header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	}

	r := &http.Request{
		Method: method,
		Header: header,
	}

	for k, v := range vars {
		r = mux.SetURLVars(r, map[string]string{k: v})
	}

	if body != "" {
		r.Body = io.NopCloser(bytes.NewBuffer([]byte(body)))
	}
	return r, user, callback
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
	e := GetDefaultEngine()
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

	e.Engine.RegisterUserController(&responseWriter, registerUserRequest)

	createdUser := builder.User{}
	builder.ParseResponse(responseWriter.Buffer.Bytes(), &createdUser)

	return &createdUser, func() {
		e.Firebase.RollbackUserRegistration(context.Background(), createdUser.FirebaseId)
	}
}
