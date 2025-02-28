package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/queries"
)

// Following these standards
// https://github.com/omniti-labs/jsend
// https://medium.com/@bojanmajed/standard-json-api-response-format-c6c1aabcaa6d

type Response struct {
	Success    bool                `json:"success"`
	Data       interface{}         `json:"data"`
	Message    string              `json:"message"`
	Pagination *queries.Pagination `json:"pagination"`
}

// NewSuccessResponse returns a Response with Success set to true.
//
// The `data` argument is returned in the `data` field of the Response.
// The `message` argument is ignored.
func newSuccessResponse(message string, data interface{}) Response {
	return Response{
		Success: true,
		Data:    data,
		Message: message,
	}
}

// newErrorResponse returns a Response with Success set to false.
//
// The `message` argument is used as the `message` field of the Response.
// The `data` argument is ignored.
func newErrorResponse(message string, data interface{}) Response {
	return Response{
		Success: false,
		Data:    data,
		Message: message,
	}
}

// SendJsonResponse writes a JSON response to the given http.ResponseWriter.
//
// It takes four arguments:
//
// - status: The HTTP status code to write to the response.
// - data: The data to include in the response body.
// - msg: A message to include in the response body.
//
// If the status code is in the 200 range, the data are included in the response body.
// If the status code is not in the 200 range, the msg is included in the response body.
//
// The function also sets the Content-Type header to "application/json", and writes the response with the given status code.
func SendJsonResponse(w http.ResponseWriter, status int, data interface{}, msg string) {
	SendJsonResponseWithPagination(w, status, data, msg, nil)
}

// SendJsonResponseWithPagination writes a JSON response to the given http.ResponseWriter.
// It takes four arguments:
//
// - status: The HTTP status code to write to the response.
// - data: The data to include in the response body.
// - msg: A message to include in the response body.
// - pagination: An optional pagination struct to include in the response body.
//
// If the status code is in the 200 range, the data and pagination are included in the response body.
// If the status code is not in the 200 range, the msg is included in the response body.
//
// The function also sets the Content-Type header to "application/json", and writes the response with the given status code.
func SendJsonResponseWithPagination(w http.ResponseWriter, status int, data interface{}, msg string, pagination *queries.Pagination) {
	var response Response

	if status >= 200 && status < 300 {
		response = newSuccessResponse(msg, data)
	} else {
		response = newErrorResponse(msg, data)
	}

	// Always include pagination if it's provided
	if pagination != nil {
		response.Pagination = pagination
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		responseBytes = []byte(fmt.Sprintf(`{"error": "%s"}`, err.Error()))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(responseBytes)
}

// ParseResponse takes a JSON byte slice and an interface to unmarshal the data into.
// The function first unmarshals the JSON into the Response struct, and then unmarshals the
// Data field of the Response struct into the provided interface. If the unmarshalling fails,
// the function returns an error.
//
// The response will have a Success field set to true if the status code of the response
// is 200, and false otherwise.
func ParseResponse(bytes []byte, v interface{}) (Response, error) {
	var response Response

	// Attempt to unmarshal the JSON bytes into the Response struct
	err := json.Unmarshal(bytes, &response)
	if err != nil {
		return response, fmt.Errorf("error unmarshalling response JSON: %w", err)
	}

	// Since the Response struct might contain a generic Data field,
	// we need to perform a two-step unmarshalling process.
	// 1. Marshal the Data field from the Response struct into a separate byte slice.
	jsonData, err := json.Marshal(response.Data)
	if err != nil {
		return response, fmt.Errorf("error marshalling response data: %w", err)
	}

	// 2. Unmarshal the marshalled data (jsonData) into the provided interface (v).
	err = json.Unmarshal(jsonData, v)
	if err != nil {
		return response, fmt.Errorf("error unmarshalling data into interface: %w", err)
	}

	// After successful unmarshalling of the data, update the Response struct with the actual data
	response.Data = v

	return response, nil
}
