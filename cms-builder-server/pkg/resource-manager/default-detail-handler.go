package resourcemanager

import (
	"net/http"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	dbQueries "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/queries"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
)

var DefaultDetailHandler ApiFunction = func(a *Resource, db *database.Database) http.HandlerFunc {
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
			SendJsonResponse(w, http.StatusForbidden, nil, "User is not allowed to access this resource")
			return
		}

		// 3. Construct Query (User Binding)
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

		msg := a.ResourceNames.Singular + " Detail"

		// 5. Send Success Response
		SendJsonResponse(w, http.StatusOK, instance, msg)
	}
}
