package file

import (
	"fmt"
	"io"
	"net/http"
	"os"

	dbQueries "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/queries"
	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	manager "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/store"
)

func DownloadStoredFileHandler(mgr *manager.ResourceManager, db *dbTypes.DatabaseConnection, st store.Store) http.HandlerFunc {

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
		if !UserIsAllowed(a.Permissions, user.GetRoles(), OperationRead, a.ResourceNames.Singular, log) {
			SendJsonResponse(w, http.StatusForbidden, nil, "User is not allowed to access this resource")
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
			SendJsonResponse(w, http.StatusNotFound, nil, "Instance not found")
			return
		}

		// Open the file
		file, err := os.Open(instance.Path)
		if err != nil {
			log.Error().Err(err).Msg("Error opening file")
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}
		defer file.Close()

		// Set headers for file download
		w.Header().Set("Content-Type", instance.MimeType)
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", instance.Name))

		// Stream the file to the response writer
		_, err = io.Copy(w, file)
		if err != nil {
			log.Error().Err(err).Msg("Error streaming file to response")
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}

	}
}
