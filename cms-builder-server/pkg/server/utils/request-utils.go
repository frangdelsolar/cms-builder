package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	authModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/models"
	loggerPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger"
	loggerTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger/types"
	svrConstants "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server/constants"
	svrTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server/types"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/utils"
	"github.com/gorilla/mux"
)

func ValidateRequestMethod(r *http.Request, method string) error {
	if r.Method != method {
		return fmt.Errorf("Method not allowed")
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
	if requestId, ok := ctx.Value(svrConstants.CtxTraceId).(string); ok {
		return requestId
	}

	return ""
}

func GetRequestLogger(r *http.Request) *loggerTypes.Logger {
	if log, ok := r.Context().Value(svrConstants.CtxRequestLogger).(*loggerTypes.Logger); ok {
		return log
	}

	return loggerPkg.Default
}

func GetRequestUser(r *http.Request) *authModels.User {
	ctxUser := r.Context().Value(svrConstants.CtxRequestUser)
	user, ok := ctxUser.(*authModels.User)
	if !ok {
		return nil
	}
	return user
}

func GetRequestIsAuth(r *http.Request) bool {
	ctxIsAuth := r.Context().Value(svrConstants.CtxRequestIsAuth)
	isAuth, ok := ctxIsAuth.(bool)
	if !ok {
		return false
	}
	return isAuth
}

type RequestContext struct {
	IsAuthenticated bool
	User            *authModels.User
	Logger          *loggerTypes.Logger
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

func GetRequestQueryParams(r *http.Request) (*svrTypes.QueryParams, error) {

	log := GetRequestLogger(r)

	params := &svrTypes.QueryParams{
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
		limit = 10
		log.Debug().Err(err).Msgf("Error validating limit. Using default limit %d", limit)
	}
	params.Limit = limit

	page, err := GetIntQueryParam("page", q)
	if err != nil {
		page = 1
		log.Debug().Err(err).Msgf("Error validating page. Using default page %d", page)
	}
	params.Page = page

	// Parse order
	orderParam := q.Get("order")
	order := ""
	order, err = ValidateOrderParam(orderParam)
	if err != nil || order == "" {
		order = "id desc"
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

func GetQueryParam(r *http.Request, param string) string {
	return r.URL.Query().Get(param)
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
