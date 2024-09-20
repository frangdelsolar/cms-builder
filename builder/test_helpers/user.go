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

func LoginUser(userData *builder.RegisterUserInput) error {
	engine := GetEngineReadyForTests()
	log, _ := engine.GetLogger()
	configReader, _ := engine.GetConfigReader()

	firebaseApiKey := configReader.GetString("firebaseApiKey")

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
		return err
	}

	// Create a new HTTP client
	client := &http.Client{}

	// Create a new request object
	r, err := http.NewRequest(http.MethodPost, firebaseLoginUrl, bytes.NewReader(bodyBytes))
	if err != nil {
		// Handle error creating the request
		log.Error().Err(err).Msg("Error creating request")
		return err
	}

	// Set the Content-Type header to application/json
	r.Header.Set("Content-Type", "application/json")

	// Send the request and get the response
	resp, err := client.Do(r)
	if err != nil {
		log.Error().Err(err).Msg("Error sending request")
		return err
	}
	defer resp.Body.Close() // Close the response body after use

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msg("Error reading response body")
		return err
	}
	log.Debug().Msg(string(data))

	// Process the response (e.g., check status code, parse the body)
	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("login failed. Status code: %d", resp.StatusCode)
		return err
	}

	// Handle successful login based on your needs (e.g., store token, use for further requests)
	return nil
}
