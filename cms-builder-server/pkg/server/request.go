package server

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/utils"
	"github.com/rs/zerolog"
)

type ContextParamKey string

const (
	CtxRequestIdentifier ContextParamKey = "requestIdentifier"
	CtxRequestStartTime  ContextParamKey = "requestStartTime"
	CtxRequestLogger     ContextParamKey = "requestLogger"
	CtxRequestIsAuth     ContextParamKey = "requestIsAuth"
	CtxRequestUser       ContextParamKey = "requestUser"
)

const (
	requestedByParamKey RequestParamKey = "requestedBy"
	rolesParamKey       RequestParamKey = "roles"
	authParamKey        RequestParamKey = "auth"
	limitParamKey       RequestParamKey = "limit"
	pageParamKey        RequestParamKey = "page"
)

type RequestParamKey string

func (r RequestParamKey) S() string {
	return string(r)
}

// ValidateRequestMethod returns an error if the request method does not match the given
// method string. The error message will include the actual request method.
func ValidateRequestMethod(r *http.Request, method string) error {
	if r.Method != method {
		return fmt.Errorf("invalid request method: %s", r.Method)
	}
	return nil
}

// GetRequestAccessToken extracts the access token from the Authorization header of the given request.
// The header should be in the format "Bearer <token>".
// If the token is not found, it returns an empty string.
func GetRequestAccessToken(r *http.Request) string {
	header := r.Header.Get("Authorization")
	if header != "" {
		tokenArray := strings.Split(header, " ")
		if len(tokenArray) == 2 && tokenArray[0] == "Bearer" {
			return tokenArray[1]
		}
	}
	return ""
}

// GetRequestId retrieves the request ID from the context of the given request.
// The request ID is set by the RequestLogMiddleware.
// If the request ID is not found, it returns an empty string.
func GetRequestId(r *http.Request) string {
	ctx := r.Context()
	if requestId, ok := ctx.Value(CtxRequestIdentifier).(string); ok {
		return requestId
	}

	return ""
}

func GetRequestLogger(r *http.Request) *zerolog.Logger {
	loggerFromContext := r.Context().Value(CtxRequestLogger)
	log, ok := loggerFromContext.(*zerolog.Logger)
	if !ok {
		return logger.Default.Logger
	}
	return log
}

func GetRequestUser(r *http.Request) *models.User {
	ctxUser := r.Context().Value(CtxRequestUser)
	user, ok := ctxUser.(*models.User)
	if !ok {
		return nil
	}
	return user
}

func GetRequestIsAuth(r *http.Request) bool {
	ctxIsAuth := r.Context().Value(CtxRequestIsAuth)
	isAuth, ok := ctxIsAuth.(bool)
	if !ok {
		return false
	}
	return isAuth
}

type RequestContext struct {
	IsAuthenticated bool
	User            *models.User
	Logger          *zerolog.Logger
	RequestId       string
}

func GetRequestContext(r *http.Request) *RequestContext {
	return &RequestContext{
		IsAuthenticated: GetRequestIsAuth(r),
		User:            GetRequestUser(r),
		Logger:          GetRequestLogger(r),
		RequestId:       GetRequestId(r),
	}
}

func GetIntQueryParam(param string, q url.Values) (int, error) {
	paramStr := q.Get(param)
	if paramStr == "" {
		return 0, fmt.Errorf("missing %s parameter", param)
	}

	paramInt, err := strconv.Atoi(paramStr)
	if err != nil {
		return 0, fmt.Errorf("invalid %s parameter", param)
	}

	return paramInt, nil
}

// QueryParams struct to hold all query parameters
type QueryParams struct {
	Limit int               `json:"limit"`
	Page  int               `json:"page"`
	Order string            `json:"order"`
	Query map[string]string `json:"query"`
}

func GetRequestQueryParams(r *http.Request) (*QueryParams, error) {

	log := GetRequestLogger(r)

	params := &QueryParams{
		Query: make(map[string]string), // Initialize the map
		Limit: 10,                      // Default limit
		Page:  1,                       // Default page
		Order: "id desc",               // Default order
	}

	var q url.Values
	if r.URL != nil {
		q = r.URL.Query()
	}

	// Parse limit (with default value and error handling)
	limit, err := GetIntQueryParam("limit", q)
	if err != nil {
		log.Error().Err(err).Msgf("Error validating limit")
		return nil, err
	}
	params.Limit = limit

	// Parse page (with default value and error handling)
	page, err := GetIntQueryParam("page", q)
	if err != nil {
		log.Error().Err(err).Msgf("Error validating page")
		return nil, err
	}
	params.Page = page

	// Parse order
	orderParam := q.Get("order")
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

func ValidateOrderParam(orderParam string) (string, error) {
	if orderParam == "" {
		return "", nil
	}

	order := ""
	fields := strings.Split(orderParam, ",")
	for _, field := range fields {

		desc := strings.HasPrefix(field, "-")

		if desc {
			field = strings.TrimPrefix(field, "-")
		}

		field = utils.SnakeCase(field)

		if desc {
			order += field + " desc,"
		} else {
			order += field + ","
		}
	}

	order = strings.TrimSuffix(order, ",")

	return order, nil
}
