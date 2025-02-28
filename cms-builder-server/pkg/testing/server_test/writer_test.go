package server_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
)

func TestLocalResponseWriter(t *testing.T) {
	// Create a LocalResponseWriter instance
	writer := &server.LocalResponseWriter{}

	t.Run("Header returns correct Content-Type", func(t *testing.T) {
		// Get the headers
		headers := writer.Header()

		// Assert that the "Content-Type" header is set to "application/json"
		assert.Equal(t, "application/json", headers.Get("Content-Type"))
	})

	t.Run("Write appends data to the buffer", func(t *testing.T) {
		// Data to write
		data := []byte(`{"message": "hello"}`)

		// Write the data to the LocalResponseWriter
		_, err := writer.Write(data)

		// Assert there was no error
		assert.NoError(t, err)

		// Get the written data
		writtenData := writer.GetWrittenData()

		// Assert that the written data matches the input data
		assert.Equal(t, `{"message": "hello"}`, writtenData)
	})

	t.Run("WriteHeader does nothing", func(t *testing.T) {
		// Use WriteHeader method with any status code
		writer.WriteHeader(200)

		// We cannot directly assert anything with WriteHeader, but we can verify that
		// it doesn't cause an error and doesn't affect the written data
		writtenData := writer.GetWrittenData()
		assert.Equal(t, "", writtenData) // No data should be written from WriteHeader
	})
}
