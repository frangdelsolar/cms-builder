package server

import (
	"fmt"
	"net/http"
)

// ValidateRequestMethod returns an error if the request method does not match the given
// method string. The error message will include the actual request method.
func ValidateRequestMethod(r *http.Request, method string) error {
	if r.Method != method {
		return fmt.Errorf("invalid request method: %s", r.Method)
	}
	return nil
}
