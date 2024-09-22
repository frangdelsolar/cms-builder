package builder

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Following these standards
// https://github.com/omniti-labs/jsend

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
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

// SendJsonResponse sends a JSON response to the client using the given http.ResponseWriter.
//
// The `status` argument is used to set the HTTP status code of the response.
// The `data` argument is used to populate the `data` field of the response.
// The `msg` argument is used to populate the `message` field of the response.
//
// If the response is for a successful status code (200), the response will have a
// `success` field set to true. For all other status codes, the response will have
// a `success` field set to false.
func SendJsonResponse(w http.ResponseWriter, status int, data interface{}, msg string) {
	var response Response

	if status >= 200 && status < 300 {
		response = newSuccessResponse(msg, data)
	} else {
		response = newErrorResponse(msg, data)
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
