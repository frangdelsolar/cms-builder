package file

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/queries"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	manager "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/store"
	"gorm.io/gorm"
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
		if !UserIsAllowed(a.Permissions, user.GetRoles(), OperationRead, a.ResourceNames.Singular, log) {
			SendJsonResponse(w, http.StatusForbidden, nil, "User is not allowed to access this resource")
			return
		}

		query := "id = ?"
		if !(a.SkipUserBinding || isAdmin) {
			query += " AND created_by_id = ?"
		}

		instanceId := GetUrlParam("id", r)
		instance := models.File{}

		res := queries.FindOne(db, &instance, instanceId, query, user.StringID())
		if res.Error != nil {
			if errors.Is(res.Error, gorm.ErrRecordNotFound) {
				SendJsonResponse(w, http.StatusNotFound, nil, "Instance not found")
				return
			}
			log.Error().Err(res.Error).Msgf("Error finding instance")
			SendJsonResponse(w, http.StatusInternalServerError, nil, res.Error.Error())
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
