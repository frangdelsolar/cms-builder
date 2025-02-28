package server_test

import (
	"net/http"
	"testing"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
	"github.com/stretchr/testify/assert"
)

func TestLocalResponseWriter_Header(t *testing.T) {
	// Create a LocalResponseWriter instance
	writer := &server.LocalResponseWriter{}

	// Call the Header method
	header := writer.Header()

	// Assertions
	assert.Equal(t, "application/json", header.Get("Content-Type"))
}

func TestLocalResponseWriter_Write(t *testing.T) {
	// Create a LocalResponseWriter instance
	writer := &server.LocalResponseWriter{}

	// Data to write
	data := []byte("test data")

	// Call the Write method
	bytesWritten, err := writer.Write(data)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, 0, bytesWritten) // The Write method always returns 0
	assert.Equal(t, "test data", writer.GetWrittenData())
}

func TestLocalResponseWriter_WriteHeader(t *testing.T) {
	// Create a LocalResponseWriter instance
	writer := &server.LocalResponseWriter{}

	// Call the WriteHeader method
	writer.WriteHeader(http.StatusOK)

	// Assertions
	// The WriteHeader method does nothing, so no explicit assertions are needed
	// This test ensures the method does not panic or cause errors
}

func TestLocalResponseWriter_GetWrittenData(t *testing.T) {
	// Create a LocalResponseWriter instance
	writer := &server.LocalResponseWriter{}

	// Write some data to the buffer
	writer.Write([]byte("test data 1"))
	writer.Write([]byte("test data 2"))

	// Call the GetWrittenData method
	writtenData := writer.GetWrittenData()

	// Assertions
	assert.Equal(t, "test data 1test data 2", writtenData)
}
