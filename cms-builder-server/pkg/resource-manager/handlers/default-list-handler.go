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

		if !(a.SkipUserBinding || isAdmin) {
			filters["created_by_id"] = user.ID
		}
		err = dbQueries.FindMany(r.Context(), log, db, instances, pagination, order, filters, []string{})
		if err != nil {
			log.Error().Err(err).Msgf("Error finding instances")
			svrUtils.SendJsonResponse(w, http.StatusNotFound, nil, "Error finding instances")
			return
		}

		// 7. Generate Success Message
		msg := a.ResourceNames.Plural + " List"

		// 8. Send Paginated Response
		svrUtils.SendJsonResponseWithPagination(w, http.StatusOK, instances, msg, pagination)
	}
}
