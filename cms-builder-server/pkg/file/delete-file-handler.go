package file

import (
	"net/http"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	dbQueries "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/queries"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	manager "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/store"
)

func DeleteStoredFilesHandler(db *database.Database, st store.Store) manager.ApiFunction {
	return func(a *manager.Resource, db *database.Database) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {

			requestCtx := GetRequestContext(r)
			log := requestCtx.Logger
			user := requestCtx.User
			requestId := requestCtx.RequestId
			isAdmin := user.HasRole(models.AdminRole)

			err := ValidateRequestMethod(r, http.MethodDelete)
			if err != nil {
				SendJsonResponse(w, http.StatusMethodNotAllowed, nil, err.Error())
				return
			}

			// 3. Check Permissions
			if !UserIsAllowed(a.Permissions, user.GetRoles(), OperationRead, a.ResourceNames.Singular, log) {
				SendJsonResponse(w, http.StatusForbidden, nil, "User is not allowed to access this resource")
				return
			}

			if !UserIsAllowed(a.Permissions, user.GetRoles(), OperationDelete, a.ResourceNames.Singular, log) {
				SendJsonResponse(w, http.StatusForbidden, nil, "User is not allowed to read this resource")
				return
			}

			filters := map[string]interface{}{
				"id": GetUrlParam("id", r),
			}

			if !(a.SkipUserBinding || isAdmin) {
				filters["created_by_id"] = user.ID
			}

			instance := models.File{}
			err = dbQueries.FindOne(r.Context(), log, db, &instance, filters)
			if err != nil {
				log.Error().Err(err).Msgf("Instance not found")
				SendJsonResponse(w, http.StatusInternalServerError, nil, "Instance not found")
				return
			}

			err = st.DeleteFile(&instance, log)
			if err != nil {
				log.Warn().Err(err).Msg("Error deleting file. Path may not exist")
			}

			err = dbQueries.Delete(r.Context(), log, db, &instance, user, requestId)
			if err != nil {
				SendJsonResponse(w, http.StatusInternalServerError, nil, "Error deleting file")
				return
			}

			SendJsonResponse(w, http.StatusOK, nil, a.ResourceNames.Singular+" deleted")
		}
	}

}
