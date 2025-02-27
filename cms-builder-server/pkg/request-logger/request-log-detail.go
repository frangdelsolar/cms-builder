package requestlogger

import (
	"errors"
	"net/http"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/queries"
	manager "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
	"gorm.io/gorm"
)

func RequestLogHandler(mgr *manager.ResourceManager, db *database.Database) http.HandlerFunc {
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

		q := "trace_id = ?" // Use parameterized query
		res := queries.FindOne(db, &instance, q, itemId)
		if res.Error != nil {
			if !errors.Is(res.Error, gorm.ErrRecordNotFound) {
				SendJsonResponse(w, http.StatusNotFound, nil, "Instance not found")
				return
			}
			log.Error().Err(res.Error).Msgf("Error finding instance")
			SendJsonResponse(w, http.StatusInternalServerError, nil, res.Error.Error())
			return
		}

		a, err = mgr.GetResource(&models.DatabaseLog{})
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
		var databaseLogs []models.DatabaseLog

		// Join HistoryEntries with RequestLog
		joinQuery := "database_logs.trace_id = ?" // Use parameterized query

		res = queries.FindMany(db, &databaseLogs, nil, "", joinQuery, itemId)
		if res.Error != nil {
			if !errors.Is(res.Error, gorm.ErrRecordNotFound) {
				log.Error().Err(res.Error).Msgf("Error finding instance")
			}
		}

		// Create a map to hold both RequestLog and HistoryEntries
		data := map[string]interface{}{
			"request_log":   instance,
			"database_logs": databaseLogs,
		}

		SendJsonResponse(w, http.StatusOK, data, "request-logs")
	}
}
