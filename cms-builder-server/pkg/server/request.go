package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/utils"
	"github.com/gorilla/mux"
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
	RequestedByParamKey RequestParamKey = "requestedBy"
	RolesParamKey       RequestParamKey = "roles"
	LimitParamKey       RequestParamKey = "limit"
	PageParamKey        RequestParamKey = "page"
)

type RequestParamKey string

func (r RequestParamKey) S() string {
	return string(r)
}

func ValidateRequestMethod(r *http.Request, method string) error {
	if r.Method != method {
		return fmt.Errorf("invalid request method: %s", r.Method)
	}
	return nil
}

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

func GetRequestId(r *http.Request) string {
	ctx := r.Context()
	if requestId, ok := ctx.Value(CtxRequestIdentifier).(string); ok {
		return requestId
	}

	return ""
}

func GetRequestLogger(r *http.Request) *logger.Logger {
	loggerFromContext := r.Context().Value(CtxRequestLogger)
	log, ok := loggerFromContext.(*logger.Logger)
	if !ok {
		fmt.Println("Logger not found in context. Returning default logger")
		return logger.Default
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
	Logger          *logger.Logger
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

	limit, err := GetIntQueryParam("limit", q)
	if err != nil {
		log.Error().Err(err).Msgf("Error validating limit")
		return nil, err
	}
	params.Limit = limit

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

func UserIsAllowed(appPermissions RolePermissionMap, userRoles []models.Role, action CrudOperation) bool {

	// Loop over the user's roles and their associated permissions
	for _, role := range userRoles {
		if _, ok := appPermissions[role]; ok {
			for _, allowedAction := range appPermissions[role] {
				if allowedAction == action {
					return true
				}
			}
		}
	}

	return false
}

func ReadRequestBody(r *http.Request) ([]byte, error) {
	if r.Body == nil {
		return []byte{}, nil
	}

	defer r.Body.Close()
	return io.ReadAll(r.Body)
}

func FormatRequestBody(r *http.Request, filterKeys map[string]bool) (map[string]interface{}, error) {
	body, err := ReadRequestBody(r)
	if err != nil {
		return map[string]interface{}{}, err
	}

	// If the body is empty, return an empty map
	if len(body) == 0 {
		return map[string]interface{}{}, nil
	}

	var unFilteredResult map[string]interface{}
	err = json.Unmarshal(body, &unFilteredResult)
	if err != nil {
		return map[string]interface{}{}, err
	}

	// Make a copy of the filter with all lowercase
	filterLowerCase := map[string]bool{}
	for key := range filterKeys {
		filterLowerCase[strings.ToLower(key)] = true
	}

	// Apply the filter to the unfiltered result
	result := make(map[string]interface{})
	for key, value := range unFilteredResult {
		lowerCaseKey := strings.ToLower(key)
		if !filterLowerCase[lowerCaseKey] {
			result[key] = value
		}
	}

	return result, nil
}

func GetUrlParam(param string, r *http.Request) string {
	return mux.Vars(r)[param]
}
