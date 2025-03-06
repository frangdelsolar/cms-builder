package handlers

import (
	"net/http"

	authConstants "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/constants"
	authUtils "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/utils"
	dbModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/models"
	dbQueries "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/queries"
	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"
	rmPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"
	svrUtils "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server/utils"
)

func TimelineHandler(m *rmPkg.ResourceManager, db *dbTypes.DatabaseConnection) http.HandlerFunc {
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
		a, err := m.GetResource(&dbModels.DatabaseLog{})
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

		// 3. Parse Query Parameters
		queryParams, err := svrUtils.GetRequestQueryParams(r)
		if err != nil {
			log.Error().Err(err).Msgf("Error validating query parameters")
			svrUtils.SendJsonResponse(w, http.StatusBadRequest, nil, err.Error())
			return
		}

		resourceName := queryParams.Query["resource_name"]
		resourceId := queryParams.Query["resource_id"]

		// 4. Verify Queried Resource
		if resourceName == "" {
			svrUtils.SendJsonResponse(w, http.StatusBadRequest, nil, "Resource Name is required")
			return
		}

		queriedApp, err := m.GetResourceByName(resourceName)
		if err != nil {
			log.Error().Err(err).Msgf("Error getting resource")
			svrUtils.SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}

		if !authUtils.UserIsAllowed(queriedApp.Permissions, user.GetRoles(), authConstants.OperationRead, resourceName, log) {
			svrUtils.SendJsonResponse(w, http.StatusForbidden, nil, "User is not allowed to read this resource")
			return
		}

		// 5. Find Query
		instances, _ := a.GetSlice()
		pagination := &dbTypes.Pagination{
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
			svrUtils.SendJsonResponse(w, http.StatusNotFound, nil, "Instance not found")
			return
		}

		svrUtils.SendJsonResponseWithPagination(w, http.StatusOK, instances, "resource timeline", pagination)

	}
}
