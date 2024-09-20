package test_helpers

import (
	"bytes"
	"net/http"
)

type ResponseWriter interface {
	Header() http.Header
	Write([]byte) (int, error)
	WriteHeader(statusCode int)
}

type MockWriter struct {
	buffer bytes.Buffer
}

func (m *MockWriter) Header() http.Header {
	return http.Header{
		"Content-Type": []string{"application/json"},
	}
}

func (m *MockWriter) Write(b []byte) (int, error) {
	m.buffer.Write(b)
	return 0, nil
}

func (m *MockWriter) WriteHeader(statusCode int) {}

func (m *MockWriter) GetWrittenData() string {
	return m.buffer.String()
}
