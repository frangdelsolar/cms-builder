package resourcemanager

import (
	"net/http"

	authConstants "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/constants"
	authUtils "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/utils"
	dbQueries "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/queries"
	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"
	rmTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager/types"
	svrUtils "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server/utils"
)

// excludedQueryKeys defines URL query parameters that should be ignored
// and not applied as simple database equality filters.
var excludedQueryKeys = map[string]bool{
	"page":  true,
	"limit": true,
	"order": true,
}

// DefaultListHandler handles the retrieval of a list of resources.
var DefaultListHandler rmTypes.ApiFunction = func(a *rmTypes.Resource, db *dbTypes.DatabaseConnection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestCtx := svrUtils.GetRequestContext(r)
		log := requestCtx.Logger
		user := requestCtx.User
		isAdmin := user.HasRole(authConstants.AdminRole)

		// 1. Validate Request Method
		err := svrUtils.ValidateRequestMethod(r, http.MethodGet)
		if err != nil {
			svrUtils.SendJsonResponse(w, http.StatusMethodNotAllowed, nil, err.Error())
			return
		}

		// 2. Check Permissions
		if !authUtils.UserIsAllowed(a.Permissions, user.GetRoles(), authConstants.OperationRead, a.ResourceNames.Singular, log) {
			svrUtils.SendJsonResponse(w, http.StatusForbidden, nil, "User is not allowed to read this resource")
			return
		}

		// 3. Parse Query Parameters
		queryParams, err := svrUtils.GetRequestQueryParams(r)
		if err != nil {
			svrUtils.SendJsonResponse(w, http.StatusBadRequest, nil, err.Error())
			return
		}

		// 4. Create Slice for Model Instances
		instances, err := a.GetSlice()
		if err != nil {
			log.Error().Err(err).Msgf("Error creating slice for model")
			svrUtils.SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}

		// 5. Construct Query and Pagination
		pagination := &dbTypes.Pagination{
			Total: 0,
			Page:  queryParams.Page,
			Limit: queryParams.Limit,
		}
		order := queryParams.Order

		filters := map[string]interface{}{}

		// 6. Apply default user binding filter (unless skipped or user is Admin)
		if !(a.SkipUserBinding || isAdmin) {
			filters["created_by_id"] = user.ID
		}

		// 7. Apply simple query filters from URL, excluding pagination/ordering keys
		for key, value := range queryParams.Query {
			// Skip parameters already handled by pagination/ordering
			if _, exists := excludedQueryKeys[key]; exists {
				continue
			}

			// Apply filter by simple equality (key = value) if value is present
			if value != "" {
				filters[key] = value
			}
		}

		// 8. Execute query
		err = dbQueries.FindMany(r.Context(), log, db, instances, pagination, order, filters, []string{})
		if err != nil {
			log.Error().Err(err).Msgf("Error finding instances")
			svrUtils.SendJsonResponse(w, http.StatusNotFound, nil, "Error finding instances")
			return
		}

		// 9. Generate Success Message
		msg := a.ResourceNames.Plural + " List"

		// 10. Send Paginated Response
		svrUtils.SendJsonResponseWithPagination(w, http.StatusOK, instances, msg, pagination)
	}
}
