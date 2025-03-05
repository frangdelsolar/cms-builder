package requestlogger

import (
	"net/http"

	dbModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/models"
	dbQueries "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/queries"
	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	manager "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
)

func RequestLogHandler(mgr *manager.ResourceManager, db *dbTypes.DatabaseConnection) http.HandlerFunc {
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

		itemId := GetUrlParam("id", r)
		instance := models.RequestLog{}

		filters := map[string]interface{}{
			"trace_id": itemId,
		}

		err = dbQueries.FindOne(r.Context(), log, db, &instance, filters)
		if err != nil {
			log.Error().Err(err).Msgf("Instance not found")
			SendJsonResponse(w, http.StatusInternalServerError, nil, "Instance not found")
			return
		}

		a, err = mgr.GetResource(&dbModels.DatabaseLog{})
		if err != nil {
			log.Error().Err(err).Msgf("Error getting resource")
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}

		if !UserIsAllowed(a.Permissions, user.GetRoles(), OperationRead, a.ResourceNames.Singular, log) {
			SendJsonResponse(w, http.StatusForbidden, nil, "User is not allowed to read this resource")
			return
		}

		// Create slice to store HistoryEntries
		var databaseLogs []dbModels.DatabaseLog

		filters = map[string]interface{}{
			"database_logs.trace_id": itemId, // join condition
		}

		err = dbQueries.FindMany(r.Context(), log, db, &databaseLogs, nil, "", filters)
		if err != nil {
			log.Warn().Err(err).Msgf("Error finding instance")
		}

		// Create a map to hold both RequestLog and HistoryEntries
		data := map[string]interface{}{
			"request_log":   instance,
			"database_logs": databaseLogs,
		}

		SendJsonResponse(w, http.StatusOK, data, "request-logs")
	}
}
