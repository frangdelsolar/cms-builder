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

var DefaultDetailHandler rmTypes.ApiFunction = func(a *rmTypes.Resource, db *dbTypes.DatabaseConnection) http.HandlerFunc {
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
			svrUtils.SendJsonResponse(w, http.StatusForbidden, nil, "User is not allowed to access this resource")
			return
		}

		// 3. Construct Query (User Binding)
		filters := map[string]interface{}{
			"id": svrUtils.GetUrlParam("id", r),
		}

		if !(a.SkipUserBinding || isAdmin) {
			filters["created_by_id"] = user.ID
		}

		instance := a.GetOne()
		err = dbQueries.FindOne(r.Context(), log, db, &instance, filters, []string{})
		if err != nil {
			log.Error().Err(err).Msgf("Instance not found")
			svrUtils.SendJsonResponse(w, http.StatusNotFound, nil, "Instance not found")
			return
		}

		msg := a.ResourceNames.Singular + " Detail"

		// 5. Send Success Response
		svrUtils.SendJsonResponse(w, http.StatusOK, instance, msg)
	}
}
