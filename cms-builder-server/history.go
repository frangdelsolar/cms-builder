package builder

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"gorm.io/gorm"
)

type CRUDAction string

const (
	CreateCRUDAction CRUDAction = "created"
	UpdateCRUDAction CRUDAction = "updated"
	DeleteCRUDAction CRUDAction = "deleted"
)

type HistoryEntry struct {
	gorm.Model
	User         *User      `json:"user"`
	UserId       string     `gorm:"foreignKey:UserId" json:"userId"`
	Username     string     `json:"username"`
	Action       CRUDAction `json:"action"`
	ResourceName string     `json:"resourceName"`
	ResourceId   string     `json:"resourceId"`
	Timestamp    string     `gorm:"type:timestamp" json:"timestamp"`
	Detail       string     `json:"detail"`
	RequestId    string     `json:"requestId"`
}

// NewLogHistoryEntry takes an action of type CRUDAction, a user ID, and an object, and returns a pointer to a HistoryEntry and an error.
// The HistoryEntry is generated by marshaling the object to JSON, and extracting the ID from it if it exists.
// The object is expected to be a struct with a JSON tag for the ID field named "ID".
// The function returns an error if the object cannot be marshaled or unmarshaled to JSON.
// The function uses the GetStructName function to get the name of the struct from the object passed in.
func NewLogHistoryEntry(action CRUDAction, user *User, object interface{}, difference interface{}, requestId string) (*HistoryEntry, error) {

	name := GetStructName(object)
	jsonData, err := json.Marshal(object)
	if err != nil {
		return nil, err
	}

	var objJson map[string]interface{}
	err = json.Unmarshal(jsonData, &objJson)
	if err != nil {
		return nil, err
	}

	resourceId := ""
	if objJson["ID"] != nil {
		resourceId = fmt.Sprintf("%v", objJson["ID"])
	}

	detail := ""
	if action == CreateCRUDAction {
		detail = string(jsonData)
	} else if action == UpdateCRUDAction {

		differenceJSON, err := json.Marshal(difference)
		if err != nil {
			return nil, err
		}

		detail = string(differenceJSON)
	}

	historyEntry := &HistoryEntry{
		Action:       action,
		UserId:       user.GetIDString(),
		Username:     user.Email,
		ResourceId:   resourceId,
		ResourceName: name,
		Timestamp:    time.Now().Format(time.RFC3339Nano),
		Detail:       detail,
		RequestId:    requestId,
	}

	return historyEntry, nil
}

// CompareInterfaces takes two objects a and b, and returns a map of their differences.
// The function is used to compare two objects and return a map of the differences.
// The returned map will have keys that are the names of the fields in the object,
// and values that are slices of two elements: the value of the field in the first object,
// and the value of the field in the second object.
func CompareInterfaces(a, b interface{}) interface{} {
	if a == nil && b == nil {
		return map[string]interface{}{}
	}

	if a == nil || b == nil {
		return map[string]interface{}{"value": []interface{}{a, b}}
	}

	aJSON, err := json.Marshal(a)
	if err != nil {
		return map[string]interface{}{"error": err.Error()}
	}

	bJSON, err := json.Marshal(b)
	if err != nil {
		return map[string]interface{}{"error": err.Error()}
	}

	var aMap map[string]interface{}
	err = json.Unmarshal(aJSON, &aMap)
	if err != nil {
		return map[string]interface{}{"error": err.Error()}
	}

	var bMap map[string]interface{}
	err = json.Unmarshal(bJSON, &bMap)
	if err != nil {
		return map[string]interface{}{"error": err.Error()}
	}

	// Initialize the result map
	res := make(map[string]interface{})

	// Iterate over all keys in the first map
	for k, va := range aMap {
		// If the key is present in the second map
		if vb, ok := bMap[k]; ok {
			// If the values are not equal, add the difference to the result map
			if !reflect.DeepEqual(va, vb) {

				// Accont for time fields
				if reflect.TypeOf(va).Kind() == reflect.String && reflect.TypeOf(vb).Kind() == reflect.String {
					timeA, errA := time.Parse(time.RFC3339Nano, va.(string)) // Parse with nanosecond precision
					timeB, errB := time.Parse(time.RFC3339Nano, vb.(string))

					if errA == nil && errB == nil { // Both are valid times
						if !timeA.Equal(timeB) { // Use time.Equal for time comparison
							res[k] = []interface{}{va, vb}
						}
					} else { // Handle parsing errors or non-time strings
						if !reflect.DeepEqual(va, vb) { //Fallback to DeepEqual
							res[k] = []interface{}{va, vb}
						}
					}
				} else {
					// If the values are both maps, recursively call GetDiff
					// Add nil map checks here
					if va != nil && vb != nil && reflect.TypeOf(va).Kind() == reflect.Map && reflect.TypeOf(vb).Kind() == reflect.Map {
						nestedDiff := CompareInterfaces(va, vb)
						// Add interface{} nil check
						if len(nestedDiff.(map[string]interface{})) > 0 {
							res[k] = nestedDiff
						}
					} else {
						res[k] = []interface{}{va, vb}
					}
				}

			}
		} else {
			// If the key is not present in the second map, add the value from the first map to the result map
			res[k] = []interface{}{va, nil}
		}
	}

	// Iterate over all keys in the second map
	for k, vb := range bMap {
		// If the key is not present in the first map, add the value from the second map to the result map
		if _, ok := aMap[k]; !ok {
			res[k] = []interface{}{nil, vb}
		}
	}

	return res
}

// GetHistoryEntryForInstanceFromDB returns a HistoryEntry if a record exists in the history table with the given parameters.
//
// The function takes a database, a user ID, a resource, a resource ID, a resource name, and a CRUD action as parameters,
// and constructs a query to retrieve a HistoryEntry from the database.
// It then executes the query and returns the HistoryEntry and any error that may have occurred.
func GetHistoryEntryForInstanceFromDB(db *Database, userId string, resource interface{}, resourceId string, resourceName string, crudAction CRUDAction) (HistoryEntry, error) {
	expectedDetail, err := json.Marshal(resource)
	if err != nil {
		return HistoryEntry{}, err
	}

	var historyEntry HistoryEntry

	q := "user_id = '" + userId + "'"
	q += " AND "
	q += "action = '" + string(crudAction) + "'"
	q += " AND "
	q += "resource_name = '" + resourceName + "'"
	q += " AND "
	q += "detail = '" + string(expectedDetail) + "'"
	q += " AND "
	q += "resource_id = '" + resourceId + "'"

	err = db.DB.First(&historyEntry).Where(q).Error
	return historyEntry, err
}

func (b *Builder) TimelineHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		err := ValidateRequestMethod(r, http.MethodGet)
		if err != nil {
			SendJsonResponse(w, http.StatusMethodNotAllowed, nil, err.Error())
			return
		}

		limit, err := strconv.Atoi(GetQueryParam("limit", r))
		if err != nil {
			log.Error().Err(err).Msgf("Error converting limit")
			limit = 10
		}

		page, err := strconv.Atoi(GetQueryParam("page", r))
		if err != nil {
			log.Error().Err(err).Msgf("Error converting page")
			page = 1
		}

		orderParam := GetQueryParam("order", r)
		order, err := ValidateOrderParam(orderParam)
		if err != nil {
			log.Error().Err(err).Msgf("Error validating order")
			log.Warn().Msg("Using default order")
		}

		a, err := b.Admin.GetApp("historyentry")
		if err != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}

		params := FormatRequestParameters(r, b)
		isAllowed := a.Permissions.HasPermission(params.Roles, OperationRead)
		if !isAllowed {
			SendJsonResponse(w, http.StatusForbidden, nil, "User is not allowed to read this resource")
			return
		}

		resourceId, err := strconv.Atoi(GetQueryParam("resource_id", r))
		if err != nil {
			SendJsonResponse(w, http.StatusBadRequest, nil, "Resource id must be an integer")
			return
		}

		resourceName := GetQueryParam("resource_name", r)
		// TODO: VALIDATE RESOURCE NAME
		if resourceName == "" {
			SendJsonResponse(w, http.StatusBadRequest, nil, "Resource name must be provided")
			return
		}

		// Create slice to store the model instances.
		instances, err := CreateSliceForUndeterminedType(a.Model)
		if err != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}

		pagination := &Pagination{
			Total: 0,
			Page:  page,
			Limit: limit,
		}

		query := "resource_id = '" + strconv.Itoa(resourceId) + "'"
		query += " AND "
		query += "resource_name = '" + resourceName + "'"

		res := b.DB.Find(instances, query, pagination, order)
		if res.Error != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, res.Error.Error())
			return
		}

		SendJsonResponseWithPagination(w, http.StatusOK, instances, a.Name()+" list", pagination)

	}
}

func (b *Builder) RequestLogHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		err := ValidateRequestMethod(r, http.MethodGet)
		if err != nil {
			SendJsonResponse(w, http.StatusMethodNotAllowed, nil, err.Error())
			return
		}

		a, err := b.Admin.GetApp("requestlog")
		if err != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}

		requestId := GetUrlParam("id", r)

		// Create instance for RequestLog
		requestLog := RequestLog{}

		query := "request_identifier = ?" // Use parameterized query
		res := b.DB.DB.Where(query, requestId).First(&requestLog)
		if res.Error != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, res.Error.Error())
			return
		}

		// Create slice to store HistoryEntries
		var historyEntries []HistoryEntry // Assuming you have a HistoryEntry struct

		// Join HistoryEntries with RequestLog
		joinQuery := "history_entries.request_id = ?" // Use parameterized query
		historyRes := b.DB.DB.Where(joinQuery, requestId).Find(&historyEntries)
		if historyRes.Error != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, historyRes.Error.Error())
			return
		}

		// Create a map to hold both RequestLog and HistoryEntries
		data := map[string]interface{}{
			"request_log":     requestLog,
			"history_entries": historyEntries,
		}

		SendJsonResponse(w, http.StatusOK, data, a.Name()+" details")
	}
}

func (b *Builder) RequestStatsHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		err := ValidateRequestMethod(r, http.MethodGet)
		if err != nil {
			SendJsonResponse(w, http.StatusMethodNotAllowed, nil, err.Error())
			return
		}

		a, err := b.Admin.GetApp("requestlog")
		if err != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}

		now := time.Now()
		oneDayAgo := now.AddDate(0, 0, -1)

		query := "timestamp > ? AND timestamp < ?"

		var statusGroupedInstances []map[string]interface{}
		statusGroupsRes := b.DB.DB.Model(a.Model).
			Select("status_code, COUNT(*) as count").
			Where(query, oneDayAgo, now).
			Group("status_code").
			Order("status_code").
			Find(&statusGroupedInstances)

		if statusGroupsRes.Error != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, statusGroupsRes.Error.Error())
			return
		}

		var methodGroupedInstances []map[string]interface{}
		methodGroupedRes := b.DB.DB.Model(a.Model).
			Select("method, COUNT(*) as count").
			Where(query, oneDayAgo, now).
			Group("method").
			Order("method").
			Find(&methodGroupedInstances)

		if methodGroupedRes.Error != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, methodGroupedRes.Error.Error())
			return
		}

		var endpointGroupedInstances []map[string]interface{}
		endpointGroupedRes := b.DB.DB.Model(a.Model).
			Select("path, COUNT(*) as count").
			Where(query, oneDayAgo, now).
			Group("path").
			Order("path").
			Find(&endpointGroupedInstances)

		if endpointGroupedRes.Error != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, endpointGroupedRes.Error.Error())
			return
		}

		var instances []map[string]interface{}
		requestLogRes := b.DB.DB.Model(a.Model).
			Select("request_identifier, timestamp, status_code, method, duration, path").
			Where(query, oneDayAgo, now).
			Order("timestamp desc").
			Find(&instances)

		if requestLogRes.Error != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, requestLogRes.Error.Error())
			return
		}

		data := map[string]interface{}{
			"endpoints":     endpointGroupedInstances,
			"method_groups": methodGroupedInstances,
			"status_groups": statusGroupedInstances,
			"requests":      instances,
		}

		SendJsonResponse(w, http.StatusOK, data, "request-logs")
	}
}
