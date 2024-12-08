package builder

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

const (
	requestedByParamKey RequestParamKey = "requested_by"
	rolesParamKey       RequestParamKey = "roles"
	authParamKey        RequestParamKey = "auth"
	limitParamKey       RequestParamKey = "limit"
	pageParamKey        RequestParamKey = "page"
)

type RequestParameters struct {
	RequestedById string
	Auth          bool
	User          *User
	Roles         []Role
}

type RequestParamKey string

func (r RequestParamKey) S() string {
	return string(r)
}

// setHeader sets a header in the HTTP request with the given parameter key and value.
// The parameter key is converted to a string using the S() method,
// and the value is set in the request header.
// If a header with the same key already exists, it will be overwritten.
func setHeader(param RequestParamKey, value string, request *http.Request) {
	request.Header.Set(param.S(), value)
}

// deleteHeader deletes a header from the HTTP request with the given parameter key.
// The parameter key is converted to a string using the S() method,
// and the header is deleted from the request.
func deleteHeader(param RequestParamKey, request *http.Request) {
	request.Header.Del(param.S())
}

// getUrlParam retrieves the value of a URL parameter from the request.
// It uses the gorilla/mux package to extract the parameter value from the URL variables.
// The function takes the parameter name as a string and the HTTP request object.
// Returns the value of the specified URL parameter as a string.
func getUrlParam(param string, r *http.Request) string {
	return mux.Vars(r)[param]
}

// getQueryParam retrieves the value of a query parameter from the request.
// The function takes the parameter name as a string and the HTTP request object.
// Returns the value of the specified query parameter as a string.
func getQueryParam(param string, r *http.Request) string {
	output := ""
	if r.URL != nil {
		output = r.URL.Query().Get(param)
	}
	return output
}

// getRequestParameters creates a RequestParameters map from the given HTTP request.
// It extracts all non-Authorization headers and query parameters from the request and
// stores them in the map.
// The requestedBy parameter is added to the map with the key "requested_by"
// The function returns the populated RequestParameters map.
func formatRequestParameters(r *http.Request, b *Builder) RequestParameters {
	params := RequestParameters{}

	user := getRequestUser(r, b)
	if user == nil {
		return params
	}

	params.User = user
	params.RequestedById = user.GetIDString()
	params.Roles = user.GetRoles()
	params.Auth = true

	return params
}

// getRequestUserId validates the access token in the Authorization header of the request.
//
// The function first retrieves the access token from the request header, then verifies it
// by calling VerifyUser on the App's admin instance. If the verification fails, it returns
// an empty string. Otherwise, it returns the ID of the verified user as a string.
func getRequestUser(r *http.Request, b *Builder) *User {
	accessToken := GetAccessTokenFromRequest(r)
	user, err := b.VerifyUser(accessToken)
	if err != nil {
		return nil
	}
	return user
}

// GetAccessTokenFromRequest extracts the access token from the Authorization header of the given request.
// The header should be in the format "Bearer <token>".
// If the token is not found, it returns an empty string.
func GetAccessTokenFromRequest(r *http.Request) string {
	header := r.Header.Get("Authorization")
	if header != "" {
		token := strings.Split(header, " ")[1]
		if token != "" {
			return token
		}
	}
	return ""
}

// readRequestBody reads the entire request body and returns the contents as a byte slice.
// It defers closing the request body until the function returns.
// It returns an error if there is a problem reading the request body.
func readRequestBody(r *http.Request) ([]byte, error) {
	defer r.Body.Close()
	return io.ReadAll(r.Body)
}

// validateRequestMethod returns an error if the request method does not match the given
// method string. The error message will include the actual request method.
func validateRequestMethod(r *http.Request, method string) error {
	if r.Method != method {
		return fmt.Errorf("invalid request method: %s", r.Method)
	}
	return nil
}
