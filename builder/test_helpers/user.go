package test_helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/frangdelsolar/cms/builder"
)

var firebaseLoginUrl = "https://www.googleapis.com/identitytoolkit/v3/relyingparty/verifyPassword?key="

// LoginUser logs in a user with the given email and password in Firebase.
//
// The function sends a request to the Firebase authentication endpoint with the email and password.
// If the request is successful, it extracts the idToken from the response and returns it. Otherwise, it
// returns an error.
//
// Parameters:
// - userData: The RegisterUserInput containing the email and password of the user to log in.
//
// Returns:
// - string: The idToken of the logged-in user.
// - error: An error if the login fails.
func LoginUser(userData *builder.RegisterUserInput) (string, error) {
	userToken := ""
	e := GetDefaultEngine()
	log, _ := e.Engine.GetLogger()
	configReader, _ := e.Engine.GetConfigReader()

	firebaseApiKey := configReader.GetString("firebaseApiKey")

	if firebaseApiKey == "" {
		log.Error().Msg("Firebase API key not set")
		return userToken, fmt.Errorf("firebase API key not set")
	}

	firebaseLoginUrl := firebaseLoginUrl + firebaseApiKey

	requestBody := map[string]string{
		"email":             userData.Email,
		"password":          userData.Password,
		"returnSecureToken": "true",
	}

	var bodyBytes []byte
	// Marshal the request body to JSON format
	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		log.Error().Err(err).Msg("Error marshalling request body")
		return userToken, err
	}

	// Create a new HTTP client
	client := &http.Client{}

	// Create a new request object
	r, err := http.NewRequest(http.MethodPost, firebaseLoginUrl, bytes.NewReader(bodyBytes))
	if err != nil {
		// Handle error creating the request
		log.Error().Err(err).Msg("Error creating request")
		return userToken, err
	}

	// Set the Content-Type header to application/json
	r.Header.Set("Content-Type", "application/json")

	// Send the request and get the response
	resp, err := client.Do(r)
	if err != nil {
		log.Error().Err(err).Msg("Error sending request")
		return userToken, err
	}
	defer resp.Body.Close() // Close the response body after use

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msg("Error reading response body")
		return userToken, err
	}

	// Unmarshal the response body
	var response map[string]interface{}
	err = json.Unmarshal(data, &response)
	if err != nil {
		log.Error().Err(err).Msg("Error unmarshalling response body")
		return userToken, err
	}

	// Extract the token from the response
	userToken, ok := response["idToken"].(string)
	if !ok {
		err := fmt.Errorf("idToken not found in response")
		return userToken, err
	}

	// Process the response (e.g., check status code, parse the body)
	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("login failed. Status code: %d", resp.StatusCode)
		return userToken, err
	}

	// Handle successful login based on your needs (e.g., store token, use for further requests)
	return userToken, nil
}

// RandomUserData returns a pointer to a builder.RegisterUserInput containing
// random data suitable for testing user registration.
func RandomUserData() *builder.RegisterUserInput {
	return &builder.RegisterUserInput{
		Name:     RandomName(),
		Email:    RandomEmail(),
		Password: "password123", // Leave all test users with the same password
	}
}
