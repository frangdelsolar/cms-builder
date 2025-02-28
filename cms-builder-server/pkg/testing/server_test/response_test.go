package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/queries"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
)

// TestSendJsonResponse_Success tests SendJsonResponse for a successful response.
func TestSendJsonResponse_Success(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"key": "value"}
	msg := "Success message"
	status := http.StatusOK

	SendJsonResponse(w, status, data, msg)

	// Verify the response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)

	// Convert response.Data to map[string]string
	responseData, ok := response.Data.(map[string]interface{})
	assert.True(t, ok, "response.Data should be of type map[string]interface{}")
	assert.Equal(t, data["key"], responseData["key"].(string)) // Compare values directly
	assert.Equal(t, msg, response.Message)
	assert.Nil(t, response.Pagination)
}

func TestSendJsonResponse_Error(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"key": "value"}
	msg := "Error message"
	status := http.StatusBadRequest

	SendJsonResponse(w, status, data, msg)

	// Verify the response
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)

	// Convert response.Data to map[string]string
	responseData, ok := response.Data.(map[string]interface{})
	assert.True(t, ok, "response.Data should be of type map[string]interface{}")
	assert.Equal(t, data["key"], responseData["key"].(string)) // Compare values directly
	assert.Equal(t, msg, response.Message)
	assert.Nil(t, response.Pagination)
}

func TestSendJsonResponseWithPagination_Success(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"key": "value"}
	msg := "Success message"
	status := http.StatusOK
	pagination := &queries.Pagination{
		Total: 100,
		Page:  1,
		Limit: 10,
	}

	SendJsonResponseWithPagination(w, status, data, msg, pagination)

	// Verify the response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)

	// Convert response.Data to map[string]string
	responseData, ok := response.Data.(map[string]interface{})
	assert.True(t, ok, "response.Data should be of type map[string]interface{}")
	assert.Equal(t, data["key"], responseData["key"].(string)) // Compare values directly
	assert.Equal(t, msg, response.Message)
	assert.Equal(t, pagination, response.Pagination)
}

func TestSendJsonResponseWithPagination_Error(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"key": "value"}
	msg := "Error message"
	status := http.StatusBadRequest
	pagination := &queries.Pagination{
		Total: 100,
		Page:  1,
		Limit: 10,
	}

	SendJsonResponseWithPagination(w, status, data, msg, pagination)

	// Verify the response
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)

	// Convert response.Data to map[string]string
	responseData, ok := response.Data.(map[string]interface{})
	assert.True(t, ok, "response.Data should be of type map[string]interface{}")
	assert.Equal(t, data["key"], responseData["key"].(string)) // Compare values directly
	assert.Equal(t, msg, response.Message)
	assert.Equal(t, pagination, response.Pagination) // Pagination should now be included
}

// TestParseResponse_Success tests ParseResponse for successful parsing of a JSON response.
func TestParseResponse_Success(t *testing.T) {
	// Create a sample JSON response
	jsonResponse := `{
		"success": true,
		"data": {"key": "value"},
		"message": "Success message"
	}`

	var data map[string]string
	response, err := ParseResponse([]byte(jsonResponse), &data)

	// Verify the parsed response
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, "Success message", response.Message)
	assert.Equal(t, map[string]string{"key": "value"}, data)

	// Dereference response.Data before comparison
	responseData, ok := response.Data.(*map[string]string)
	assert.True(t, ok, "response.Data should be of type *map[string]string")
	assert.Equal(t, data, *responseData)
}

// TestParseResponse_Error tests ParseResponse for error handling with invalid JSON.
func TestParseResponse_Error(t *testing.T) {
	// Create an invalid JSON response
	invalidJSON := `{invalid json}`

	var data map[string]string
	_, err := ParseResponse([]byte(invalidJSON), &data)

	// Verify the error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error unmarshalling response JSON")
}

// TestParseResponse_DataError tests ParseResponse for error handling when unmarshalling the data field fails.
func TestParseResponse_DataError(t *testing.T) {
	// Create a JSON response with invalid data
	jsonResponse := `{
		"success": true,
		"data": "invalid data",
		"message": "Success message"
	}`

	var data map[string]string
	_, err := ParseResponse([]byte(jsonResponse), &data)

	// Verify the error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error unmarshalling data into interface")
}
