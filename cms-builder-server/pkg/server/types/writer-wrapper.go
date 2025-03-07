package server

import (
	"bytes"
	"net/http"
)

type WriterWrapper struct {
	http.ResponseWriter
	StatusCode int
	Body       *bytes.Buffer
}

func (w *WriterWrapper) WriteHeader(statusCode int) {
	w.StatusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *WriterWrapper) Write(b []byte) (int, error) {
	if w.Body != nil {
		w.Body.Write(b)
	}
	return w.ResponseWriter.Write(b)
}
