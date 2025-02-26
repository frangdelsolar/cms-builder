package databaselogger

import (
	"net/http"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/queries"
	manager "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
)

func TimelineHandler(m *manager.ResourceManager, db *database.Database) http.HandlerFunc {
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
		a, err := m.GetResource(&models.DatabaseLog{})
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

		// 3. Parse Query Parameters
		queryParams, err := GetRequestQueryParams(r)
		if err != nil {
			log.Error().Err(err).Msgf("Error validating query parameters")
			SendJsonResponse(w, http.StatusBadRequest, nil, err.Error())
			return
		}

		resourceName := queryParams.Query["resource_name"]
		resourceId := queryParams.Query["resource_id"]

		// 4. Verify Queried Resource
		if resourceName == "" {
			SendJsonResponse(w, http.StatusBadRequest, nil, "Resource Name is required")
			return
		}

		queriedApp, err := m.GetResourceByName(resourceName)
		if err != nil {
			log.Error().Err(err).Msgf("Error getting resource")
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}

		if !UserIsAllowed(queriedApp.Permissions, user.GetRoles(), OperationRead, resourceName, log) {
			SendJsonResponse(w, http.StatusForbidden, nil, "User is not allowed to read this resource")
			return
		}

		// 5. Find Query
		instance := a.GetOne()
		query := "resource_id = ? AND resource_name = ?"

		res := queries.FindOne(db, instance, query, resourceId, resourceName)
		if res.Error != nil {
			log.Error().Err(res.Error).Msgf("Error finding instance")
			SendJsonResponse(w, http.StatusInternalServerError, nil, res.Error.Error())
			return
		}

		if instance == nil {
			log.Error().Msgf("Instance not found")
			SendJsonResponse(w, http.StatusNotFound, nil, "Instance not found")
			return
		}

		SendJsonResponse(w, http.StatusOK, instance, "timeline-detail")

	}
}
