package file

import (
	"net/http"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/queries"
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

			query := ""
			if !(a.SkipUserBinding || isAdmin) {
				query = "created_by_id = ?"
			}

			instanceId := GetUrlParam("id", r)
			instance := models.File{}
			res := queries.FindOne(db, &instance, instanceId, query, user.StringID())

			if res.Error != nil {
				log.Error().Err(res.Error).Msgf("Error finding instance")
				SendJsonResponse(w, http.StatusInternalServerError, nil, res.Error.Error())
				return
			}

			if instance == (models.File{}) {
				log.Error().Msgf("Instance not found")
				SendJsonResponse(w, http.StatusNotFound, nil, "Instance not found")
				return
			}

			err = st.DeleteFile(&instance)
			if err != nil {
				SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
				return
			}

			res = queries.Delete(db, &instance, user, requestId)
			if res.Error != nil {
				SendJsonResponse(w, http.StatusInternalServerError, nil, res.Error.Error())
				return
			}

			SendJsonResponse(w, http.StatusOK, nil, a.ResourceNames.Singular+" deleted")
		}
	}

}
