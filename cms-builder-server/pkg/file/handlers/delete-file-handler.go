package file

import (
	"net/http"

	authConstants "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/constants"
	authUtils "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/utils"
	dbQueries "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/queries"
	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"
	fileModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/file/models"
	rmTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager/types"
	svrUtils "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server/utils"
	storeTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/store/types"
)

func DeleteStoredFilesHandler(db *dbTypes.DatabaseConnection, st storeTypes.Store) rmTypes.ApiFunction {
	return func(a *rmTypes.Resource, db *dbTypes.DatabaseConnection) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {

			requestCtx := svrUtils.GetRequestContext(r)
			log := requestCtx.Logger
			user := requestCtx.User
			requestId := requestCtx.RequestId
			isAdmin := user.HasRole(authConstants.AdminRole)

			err := svrUtils.ValidateRequestMethod(r, http.MethodDelete)
			if err != nil {
				svrUtils.SendJsonResponse(w, http.StatusMethodNotAllowed, nil, err.Error())
				return
			}

			// 3. Check Permissions
			if !authUtils.UserIsAllowed(a.Permissions, user.GetRoles(), authConstants.OperationRead, a.ResourceNames.Singular, log) {
				svrUtils.SendJsonResponse(w, http.StatusForbidden, nil, "User is not allowed to access this resource")
				return
			}

			if !authUtils.UserIsAllowed(a.Permissions, user.GetRoles(), authConstants.OperationDelete, a.ResourceNames.Singular, log) {
				svrUtils.SendJsonResponse(w, http.StatusForbidden, nil, "User is not allowed to read this resource")
				return
			}

			filters := map[string]interface{}{
				"id": svrUtils.GetUrlParam("id", r),
			}

			if !(a.SkipUserBinding || isAdmin) {
				filters["created_by_id"] = user.ID
			}

			instance := fileModels.File{}
			err = dbQueries.FindOne(r.Context(), log, db, &instance, filters, []string{})
			if err != nil {
				log.Error().Err(err).Msgf("Instance not found")
				svrUtils.SendJsonResponse(w, http.StatusInternalServerError, nil, "Instance not found")
				return
			}

			err = st.DeleteFile(&instance, log)
			if err != nil {
				log.Warn().Err(err).Msg("Error deleting file. Path may not exist")
			}

			err = dbQueries.Delete(r.Context(), log, db, &instance, user, requestId)
			if err != nil {
				svrUtils.SendJsonResponse(w, http.StatusInternalServerError, nil, "Error deleting file")
				return
			}

			svrUtils.SendJsonResponse(w, http.StatusOK, nil, a.ResourceNames.Singular+" deleted")
		}
	}

}
