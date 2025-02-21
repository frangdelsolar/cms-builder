package builder

import (
	"bytes"
	"net/http"
)

type ResponseWriter interface {
	Header() http.Header
	Write([]byte) (int, error)
	WriteHeader(statusCode int)
}

type LocalResponseWriter struct {
	Buffer bytes.Buffer
}

// Header returns a Header that can be used as a response header.
// The returned Header contains one entry: "Content-Type" set to "application/json".
func (m *LocalResponseWriter) Header() http.Header {
	return http.Header{
		"Content-Type": []string{"application/json"},
	}
}

// Write writes the given []byte to the internal buffer.
// It returns 0 and nil because it is not possible to write to the buffer in a
// way that would return an error.
func (m *LocalResponseWriter) Write(b []byte) (int, error) {
	m.Buffer.Write(b)
	return 0, nil
}

func (m *LocalResponseWriter) WriteHeader(statusCode int) {}

// GetWrittenData returns the data written to the MockWriter as a string.
func (m *LocalResponseWriter) GetWrittenData() string {
	return m.Buffer.String()
}
