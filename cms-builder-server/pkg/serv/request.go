package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

const (
	requestedByParamKey RequestParamKey = "requestedBy"
	rolesParamKey       RequestParamKey = "roles"
	authParamKey        RequestParamKey = "auth"
	limitParamKey       RequestParamKey = "limit"
	pageParamKey        RequestParamKey = "page"
)

type RequestParameters struct {
	RequestId     string
	RequestedById string
	Auth          bool
	User          *User
	Roles         []Role
}

type RequestParamKey string

func (r RequestParamKey) S() string {
	return string(r)
}

// SetHeader sets a header in the HTTP request with the given parameter key and value.
// The parameter key is converted to a string using the S() method,
// and the value is set in the request header.
// If a header with the same key already exists, it will be overwritten.
func SetHeader(param RequestParamKey, value string, request *http.Request) {
	request.Header.Set(param.S(), value)
}

// DeleteHeader deletes a header from the HTTP request with the given parameter key.
// The parameter key is converted to a string using the S() method,
// and the header is deleted from the request.
func DeleteHeader(param RequestParamKey, request *http.Request) {
	request.Header.Del(param.S())
}

// GetUrlParam retrieves the value of a URL parameter from the request.
// It uses the gorilla/mux package to extract the parameter value from the URL variables.
// The function takes the parameter name as a string and the HTTP request object.
// Returns the value of the specified URL parameter as a string.
func GetUrlParam(param string, r *http.Request) string {
	return mux.Vars(r)[param]
}

// QueryParams struct to hold all query parameters
type QueryParams struct {
	Limit int               `json:"limit"`
	Page  int               `json:"page"`
	Order string            `json:"order"`
	Query map[string]string `json:"query"`
}

func GetQueryParams(r *http.Request) (*QueryParams, error) {
	params := &QueryParams{
		Query: make(map[string]string), // Initialize the map
		Limit: 10,                      // Default limit
		Page:  1,                       // Default page
	}

	var q url.Values
	if r.URL != nil {
		q = r.URL.Query()
	}

	// Parse limit (with default value and error handling)
	limitStr := q.Get("limit")
	if limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			return nil, err // Return error if limit is not a number
		}
		params.Limit = limit
	}

	// Parse page (with default value and error handling)
	pageStr := q.Get("page")
	if pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			return nil, err // Return error if page is not a number
		}
		params.Page = page
	}

	orderParam := GetQueryParam("order", r)
	order, err := ValidateOrderParam(orderParam)
	if err != nil {
		log.Error().Err(err).Msgf("Error validating order")
		log.Warn().Msg("Using default order")
	}

	params.Order = order

	// Parse query parameters into the map
	for key, values := range q {
		if key != "limit" && key != "page" && key != "order" { // Exclude standard params
			params.Query[key] = strings.Join(values, ",") // Assuming only one value per query parameter for now. Can be modified to handle multiple values per key if needed.
		}
	}

	return params, nil
}

// GetQueryParam retrieves the value of a query parameter from the request.
// The function takes the parameter name as a string and the HTTP request object.
// Returns the value of the specified query parameter as a string.
func GetQueryParam(param string, r *http.Request) string {
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
func FormatRequestParameters(r *http.Request, b *Builder) RequestParameters {
	params := RequestParameters{}

	user := GetRequestUser(r, b)
	if user == nil {
		return params
	}

	params.User = user
	params.RequestedById = user.GetIDString()
	params.Roles = user.GetRoles()
	params.Auth = true
	params.RequestId = GetRequestId(r)

	return params
}

func GetRequestId(r *http.Request) string {

	// get from context requestIdentifier
	ctx := r.Context()
	if requestId, ok := ctx.Value("requestIdentifier").(string); ok {
		return requestId
	}

	return ""
}

// getRequestUserId validates the access token in the Authorization header of the request.
//
// The function first retrieves the access token from the request header, then verifies it
// by calling VerifyUser on the App's admin instance. If the verification fails, it returns
// an empty string. Otherwise, it returns the ID of the verified user as a string.
func GetRequestUser(r *http.Request, b *Builder) *User {

	// TODO: should pass user through context instead of verifying agains, as this has been done in the usermiddleware
	godToken := r.Header.Get(GodTokenHeader)
	accessToken := GetAccessTokenFromRequest(r)
	requestId := GetRequestId(r)

	var localUser *User
	if godToken != "" {
		localUser, _ = b.VerifyGodUser(godToken)
	} else {
		localUser, _ = b.VerifyUser(accessToken, requestId)
	}

	return localUser
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

// ReadRequestBody reads the entire request body and returns the contents as a byte slice.
// It defers closing the request body until the function returns.
// It returns an error if there is a problem reading the request body.
func ReadRequestBody(r *http.Request) ([]byte, error) {
	if r.Body == nil {
		return []byte{}, nil
	}

	defer r.Body.Close()
	return io.ReadAll(r.Body)
}

// FormatRequestBody reads the request body and filters out any keys specified in the filterKeys map.
// It returns the filtered request body as a map[string]interface{}.
// If there is an error reading the request body, it returns an empty map.
// The function applies the filter with a case-insensitive comparison.
func FormatRequestBody(r *http.Request, filterKeys map[string]bool) (map[string]interface{}, error) {
	var unFilteredResult map[string]interface{}
	body, err := ReadRequestBody(r)
	if err != nil {
		return map[string]interface{}{}, err
	}

	err = json.Unmarshal(body, &unFilteredResult)
	if err != nil {
		return map[string]interface{}{}, err
	}

	// make a copy of the filter with all lowercase
	filterLowerCase := map[string]bool{}
	for key := range filterKeys {
		filterLowerCase[strings.ToLower(key)] = true
	}

	// apply the filter to the unfiltered result
	result := make(map[string]interface{})
	for key, value := range unFilteredResult {
		lowerCaseKey := strings.ToLower(key)
		if !filterLowerCase[lowerCaseKey] {
			result[key] = value
		}
	}

	return result, nil
}

// ValidateRequestMethod returns an error if the request method does not match the given
// method string. The error message will include the actual request method.
func ValidateRequestMethod(r *http.Request, method string) error {
	if r.Method != method {
		return fmt.Errorf("invalid request method: %s", r.Method)
	}
	return nil
}
