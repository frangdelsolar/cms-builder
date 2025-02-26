package resourcemanager

import (
	"net/http"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/queries"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
)

// DefaultDeleteHandler handles the deletion of a resource.
var DefaultDeleteHandler ApiFunction = func(a *Resource, db *database.Database) http.HandlerFunc {
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

		// 3. Construct Query (User Binding)
		query := ""
		if !(a.SkipUserBinding || isAdmin) {
			query = "created_by_id = ?"
		}

		// 4. Retrieve Instance ID and Fetch Instance
		instanceId := GetUrlParam("id", r)
		instance := a.GetOne()
		res := queries.FindOne(db, instance, instanceId, query, user.StringID())

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

		// 5. Delete Instance
		res = queries.Delete(db, instance, user, requestId)
		if res.Error != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, res.Error.Error())
			return
		}

		// 6. Generate Success Message
		msg := a.ResourceNames.Singular + " has been eleted"

		// 7. Send Success Response
		SendJsonResponse(w, http.StatusOK, nil, msg)
	}
}
