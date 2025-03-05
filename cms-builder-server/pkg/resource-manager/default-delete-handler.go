package resourcemanager

import (
	"net/http"

	dbQueries "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/queries"
	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
)

// DefaultDeleteHandler handles the deletion of a resource.
var DefaultDeleteHandler ApiFunction = func(a *Resource, db *dbTypes.DatabaseConnection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestCtx := GetRequestContext(r)
		log := requestCtx.Logger
		user := requestCtx.User
		requestId := requestCtx.RequestId
		isAdmin := user.HasRole(models.AdminRole)

		// 1. Validate Request Method
		err := ValidateRequestMethod(r, http.MethodDelete)
		if err != nil {
			SendJsonResponse(w, http.StatusMethodNotAllowed, nil, err.Error())
			return
		}

		// 2. Check Permissions
		if !UserIsAllowed(a.Permissions, user.GetRoles(), OperationRead, a.ResourceNames.Singular, log) {
			SendJsonResponse(w, http.StatusForbidden, nil, "User is not allowed to access this resource")
			return
		}

		if !UserIsAllowed(a.Permissions, user.GetRoles(), OperationDelete, a.ResourceNames.Singular, log) {
			SendJsonResponse(w, http.StatusForbidden, nil, "User is not allowed to delete this resource")
			return
		}

		filters := map[string]interface{}{
			"id": GetUrlParam("id", r),
		}

		if !(a.SkipUserBinding || isAdmin) {
			filters["created_by_id"] = user.ID
		}

		instance := a.GetOne()
		err = dbQueries.FindOne(r.Context(), log, db, &instance, filters)
		if err != nil {
			log.Error().Err(err).Msgf("Instance not found")
			SendJsonResponse(w, http.StatusNotFound, nil, "Instance not found")
			return
		}

		// 5. Delete Instance
		err = dbQueries.Delete(r.Context(), log, db, instance, user, requestId)
		if err != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, "Error deleting resource")
			return
		}

		// 6. Generate Success Message
		msg := a.ResourceNames.Singular + " has been deleted"

		// 7. Send Success Response
		SendJsonResponse(w, http.StatusOK, nil, msg)
	}
}
