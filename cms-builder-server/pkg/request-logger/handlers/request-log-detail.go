package handlers

import (
	"net/http"

	authConstants "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/constants"
	authUtils "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/utils"
	dbModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/models"
	dbQueries "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/queries"
	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"
	rlModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/request-logger/models"
	rmPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"
	svrUtils "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server/utils"
)

func RequestLogHandler(mgr *rmPkg.ResourceManager, db *dbTypes.DatabaseConnection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		requestCtx := svrUtils.GetRequestContext(r)
		log := requestCtx.Logger
		user := requestCtx.User

		// 1. Validate Request Method
		err := svrUtils.ValidateRequestMethod(r, http.MethodGet)
		if err != nil {
			log.Error().Err(err).Msgf("Error validating request method")
			svrUtils.SendJsonResponse(w, http.StatusMethodNotAllowed, nil, err.Error())
			return
		}

		// 2. Get Resource
		a, err := mgr.GetResource(&rlModels.RequestLog{})
		if err != nil {
			log.Error().Err(err).Msgf("Error getting resource")
			svrUtils.SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}

		// 3. Check Permissions
		if !authUtils.UserIsAllowed(a.Permissions, user.GetRoles(), authConstants.OperationRead, a.ResourceNames.Singular, log) {
			svrUtils.SendJsonResponse(w, http.StatusForbidden, nil, "User is not allowed to read this resource")
			return
		}

		itemId := svrUtils.GetUrlParam("id", r)
		instance := rlModels.RequestLog{}

		filters := map[string]interface{}{
			"trace_id": itemId,
		}

		err = dbQueries.FindOne(r.Context(), log, db, &instance, filters, []string{})
		if err != nil {
			log.Error().Err(err).Msgf("Instance not found")
			svrUtils.SendJsonResponse(w, http.StatusInternalServerError, nil, "Instance not found")
			return
		}

		a, err = mgr.GetResource(&dbModels.DatabaseLog{})
		if err != nil {
			log.Error().Err(err).Msgf("Error getting resource")
			svrUtils.SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}

		if !authUtils.UserIsAllowed(a.Permissions, user.GetRoles(), authConstants.OperationRead, a.ResourceNames.Singular, log) {
			svrUtils.SendJsonResponse(w, http.StatusForbidden, nil, "User is not allowed to read this resource")
			return
		}

		// Create slice to store HistoryEntries
		var databaseLogs []dbModels.DatabaseLog

		filters = map[string]interface{}{
			"database_logs.trace_id": itemId, // join condition
		}

		err = dbQueries.FindMany(r.Context(), log, db, &databaseLogs, nil, "", filters, []string{})
		if err != nil {
			log.Warn().Err(err).Msgf("Error finding instance")
		}

		// Create a map to hold both RequestLog and HistoryEntries
		data := map[string]interface{}{
			"request_log":   instance,
			"database_logs": databaseLogs,
		}

		svrUtils.SendJsonResponse(w, http.StatusOK, data, "request-logs")
	}
}
