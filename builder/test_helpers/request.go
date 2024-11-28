package test_helpers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/frangdelsolar/cms/builder"
	"github.com/gorilla/mux"
)

// NewRequest creates a new HTTP request with the specified method and body.
// It optionally authenticates the request using the provided user. The function
// returns the created request, the user, and a rollback function to undo any
// user-related changes.
//
// Parameters:
// - method: The HTTP method (e.g., "GET", "POST").
// - body: The request body as a string.
// - authenticate: A boolean indicating whether to authenticate the request.
// - user: A pointer to a builder.User for authentication purposes.
// - vars: A map of variables to include in the request URL.
//
// Returns:
// - *http.Request: The newly created HTTP request.
// - *builder.User: The user used for authentication.
// - func(): A function to roll back user-related changes.
func NewRequest(method string, body string, authenticate bool, user *builder.User, vars map[string]string) (*http.Request, *builder.User, func()) {
	return NewRequestWithFile(method, body, "", authenticate, user, vars)
}

// NewRequestWithFile creates a new *http.Request with a JSON body and an optional file attachment.
// The request can be authenticated with a test user, which is created and rolled back when the callback
// function is called.
// The request body is set to the provided JSON string, and the file is attached as a form field named "file".
// The Content-Type header is set to application/json, and the Authorization header is set to a valid
// Bearer token if authentication is required.
// The function returns the new request, the test user, and a callback function that should be called
// to clean up the test user when it is no longer needed.
func NewRequestWithFile(method string, body string, filePath string, authenticate bool, user *builder.User, vars map[string]string) (*http.Request, *builder.User, func()) {
	callback := func() {}

	header := http.Header{
		"Content-Type": []string{"application/json"},
	}

	// Validate file existence
	if filePath != "" {
		_, err := os.Stat(filePath)
		if err != nil {
			if os.IsNotExist(err) {
				panic(fmt.Errorf("file '%s' does not exist", filePath))
			} else {
				panic(err)
			}
		}

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

	if filePath != "" {
		// Create a new multipart writer
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		// Create a form field for the file
		file, err := os.Open(filePath)
		if err != nil {
			panic(err)
		}
		defer file.Close()

		fw, err := writer.CreateFormFile("file", filepath.Base(file.Name()))
		if err != nil {
			panic(err)
		}
		_, err = io.Copy(fw, file)
		if err != nil {
			panic(err)
		}

		// Close the multipart writer
		err = writer.Close()
		if err != nil {
			panic(err)
		}

		// Set the Content-Type header
		r.Header.Set("Content-Type", writer.FormDataContentType())

		// Set the request body
		r.Body = ioutil.NopCloser(body)

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
	e, err := GetDefaultEngine()
	if err != nil {
		panic(err)
	}

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
