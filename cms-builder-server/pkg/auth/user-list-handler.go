package auth

import (
	"net/http"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/queries"
	dbQueries "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/queries"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	mgr "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
)

// DefaultListHandler handles the retrieval of a list of resources.
var UserListHandler mgr.ApiFunction = func(a *mgr.Resource, db *database.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestCtx := GetRequestContext(r)
		log := requestCtx.Logger
		user := requestCtx.User
		isAdmin := user.HasRole(models.AdminRole)

		// 1. Validate Request Method
		err := ValidateRequestMethod(r, http.MethodGet)
		if err != nil {
			SendJsonResponse(w, http.StatusMethodNotAllowed, nil, err.Error())
			return
		}

		// 2. Check Permissions
		if !UserIsAllowed(a.Permissions, user.GetRoles(), OperationRead, a.ResourceNames.Singular, log) {
			SendJsonResponse(w, http.StatusForbidden, nil, "User is not allowed to read this resource")
			return
		}

		// 3. Parse Query Parameters
		queryParams, err := GetRequestQueryParams(r)
		if err != nil {
			SendJsonResponse(w, http.StatusBadRequest, nil, err.Error())
			return
		}

		// 4. Create Slice for Model Instances
		instances, err := a.GetSlice()
		if err != nil {
			log.Error().Err(err).Msgf("Error creating slice for model")
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}

		// 5. Construct Query and Pagination
		pagination := &queries.Pagination{
			Total: 0,
			Page:  queryParams.Page,
			Limit: queryParams.Limit,
		}
		order := queryParams.Order

		filters := map[string]interface{}{}
		if !isAdmin {
			filters["id"] = user.ID
		}

		fbId, ok := queryParams.Query["fbId"]
		if ok {
			filters["firebase_id"] = fbId
		}

		err = dbQueries.FindMany(r.Context(), log, db, instances, pagination, order, filters)
		if err != nil {
			log.Error().Err(err).Msgf("Error finding instances")
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}

		// 7. Generate Success Message
		msg := a.ResourceNames.Plural + " List"

		// 8. Send Paginated Response
		SendJsonResponseWithPagination(w, http.StatusOK, instances, msg, pagination)
	}
}
