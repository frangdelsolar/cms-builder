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

func DownloadStoredFileHandler(mgr *manager.ResourceManager, db *database.Database, st store.Store) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method == http.MethodHead {
			w.Header().Set("Content-Type", "bytes")
			w.WriteHeader(http.StatusOK)
			return
		}

		requestCtx := GetRequestContext(r)
		log := requestCtx.Logger
		user := requestCtx.User
		isAdmin := user.HasRole(models.AdminRole)

		err := ValidateRequestMethod(r, http.MethodGet)
		if err != nil {
			SendJsonResponse(w, http.StatusMethodNotAllowed, nil, err.Error())
			return
		}

		a, err := mgr.GetResource(models.File{})
		if err != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}

		// 3. Check Permissions
		if !UserIsAllowed(a.Permissions, user.GetRoles(), OperationRead) {
			SendJsonResponse(w, http.StatusForbidden, nil, "User is not allowed to access this resource")
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

		bytes, err := st.ReadFile(&instance)
		if err != nil {
			log.Error().Err(err).Msg("Error reading file")
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}

		w.Header().Set("Content-Type", "bytes")
		w.Write(bytes)
	}
}
