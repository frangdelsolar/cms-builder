package databaselogger

import (
	"net/http"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/queries"
	dbQueries "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/queries"
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
		a, err := m.GetResource(&database.DatabaseLog{})
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
		instances, _ := a.GetSlice()
		pagination := &queries.Pagination{
			Total: 0,
			Page:  queryParams.Page,
			Limit: queryParams.Limit,
		}
		filters := map[string]interface{}{
			"resource_id":   resourceId,
			"resource_name": resourceName,
		}

		err = dbQueries.FindMany(r.Context(), log, db, instances, pagination, queryParams.Order, filters)
		if err != nil {
			SendJsonResponse(w, http.StatusNotFound, nil, "Instance not found")
			return
		}

		SendJsonResponseWithPagination(w, http.StatusOK, instances, "resource timeline", pagination)

	}
}
