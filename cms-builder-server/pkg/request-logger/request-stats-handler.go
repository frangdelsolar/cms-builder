package requestlogger

import (
	"net/http"
	"time"
)

func RequestStatsHandler() http.HandlerFunc {
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

		var userGroupedInstances []map[string]interface{}
		userGroupedRes := b.DB.DB.Model(a.Model).
			Select("users.email, COUNT(*) as count").
			Joins("JOIN users ON users.id = request_logs.user_id"). // Join with the users table
			Where(query, oneDayAgo, now).
			Group("users.email"). // Group by email
			Order("users.email"). // Order by email
			Find(&userGroupedInstances)

		if userGroupedRes.Error != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, userGroupedRes.Error.Error())
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
			"users":         userGroupedInstances,
			"endpoints":     endpointGroupedInstances,
			"method_groups": methodGroupedInstances,
			"status_groups": statusGroupedInstances,
			"requests":      instances,
		}

		SendJsonResponse(w, http.StatusOK, data, "request-logs")
	}
}
