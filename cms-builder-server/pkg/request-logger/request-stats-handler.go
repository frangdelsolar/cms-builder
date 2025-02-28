package requestlogger

import (
	"net/http"
	"time"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	manager "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
)

func RequestStatsHandler(mgr *manager.ResourceManager, db *database.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		requestCtx := GetRequestContext(r)
		log := requestCtx.Logger
		user := requestCtx.User

		// 1. Validate Request Method
		err := ValidateRequestMethod(r, http.MethodGet)
		if err != nil {
			log.Error().Err(err).Msgf("Error validating request method")
			SendJsonResponse(w, http.StatusMethodNotAllowed, nil, err.Error())
			return
		}

		// 2. Get Resource
		a, err := mgr.GetResource(&models.RequestLog{})
		if err != nil {
			log.Error().Err(err).Msgf("Error getting resource")
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}

		// 3. Check Permissions
		if !UserIsAllowed(a.Permissions, user.GetRoles(), OperationRead, a.ResourceNames.Singular, log) {
			SendJsonResponse(w, http.StatusForbidden, nil, "User is not allowed to read this resource")
			return
		}

		now := time.Now()
		oneDayAgo := now.AddDate(0, 0, -1)

		query := "timestamp > ? AND timestamp < ?"

		var statusGroupedInstances []map[string]interface{}
		statusGroupsRes := db.DB.Model(a.Model).
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
		methodGroupedRes := db.DB.Model(a.Model).
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
		userGroupedRes := db.DB.Model(a.Model).
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
		endpointGroupedRes := db.DB.Model(a.Model).
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
		requestLogRes := db.DB.Model(a.Model).
			Select("trace_id, timestamp, status_code, method, duration, path").
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
