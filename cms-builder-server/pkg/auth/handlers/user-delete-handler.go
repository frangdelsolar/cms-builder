package auth

import (
	"net/http"

	authConstants "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/constants"
	authUtils "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/utils"
	dbQueries "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/queries"
	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"
	rmTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager/types"
	svrUtils "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server/utils"
)

// DefaultDeleteHandler handles the deletion of a resource.
var UserDeleteHandler rmTypes.ApiFunction = func(a *rmTypes.Resource, db *dbTypes.DatabaseConnection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestCtx := svrUtils.GetRequestContext(r)
		log := requestCtx.Logger
		user := requestCtx.User
		requestId := requestCtx.RequestId
		isAdmin := user.HasRole(authConstants.AdminRole)

		// 1. Validate Request Method
		err := svrUtils.ValidateRequestMethod(r, http.MethodDelete)
		if err != nil {
			svrUtils.SendJsonResponse(w, http.StatusMethodNotAllowed, nil, err.Error())
			return
		}

		// 2. Check Permissions
		if !authUtils.UserIsAllowed(a.Permissions, user.GetRoles(), authConstants.OperationRead, a.ResourceNames.Singular, log) {
			svrUtils.SendJsonResponse(w, http.StatusForbidden, nil, "User is not allowed to access this resource")
			return
		}

		if !authUtils.UserIsAllowed(a.Permissions, user.GetRoles(), authConstants.OperationDelete, a.ResourceNames.Singular, log) {
			svrUtils.SendJsonResponse(w, http.StatusForbidden, nil, "User is not allowed to delete this resource")
			return
		}

		if !isAdmin {
			if user.StringID() != svrUtils.GetUrlParam("id", r) {
				svrUtils.SendJsonResponse(w, http.StatusNotFound, nil, "Instance not found")
				return
			}
		}

		filters := map[string]interface{}{
			"id": svrUtils.GetUrlParam("id", r),
		}

		instance := a.GetOne()
		err = dbQueries.FindOne(r.Context(), log, db, &instance, filters, []string{})
		if err != nil {
			log.Error().Err(err).Msgf("Instance not found")
			svrUtils.SendJsonResponse(w, http.StatusNotFound, nil, "Instance not found")
			return
		}

		// 5. Delete Instance
		err = dbQueries.Delete(r.Context(), log, db, instance, user, requestId)
		if err != nil {
			svrUtils.SendJsonResponse(w, http.StatusInternalServerError, nil, "Error deleting resource")
			return
		}

		// 6. Generate Success Message
		msg := a.ResourceNames.Singular + " has been deleted"

		// 7. Send Success Response
		svrUtils.SendJsonResponse(w, http.StatusOK, nil, msg)
	}
}
