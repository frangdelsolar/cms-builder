package resourcemanager

import (
	"errors"
	"net/http"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/queries"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
	"gorm.io/gorm"
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
		query := "id = ?"
		if !(a.SkipUserBinding || isAdmin) {
			query += " AND created_by_id = ?"
		}

		// 4. Retrieve Instance ID and Fetch Instance
		instanceId := GetUrlParam("id", r)
		instance := a.GetOne()

		log.Debug().Str("path", r.URL.Path).Str("instance_id", instanceId).Str("q", query).Msg("Fetching Instance")

		res := queries.FindOne(db, instance, query, instanceId, user.StringID())
		if res.Error != nil {
			if errors.Is(res.Error, gorm.ErrRecordNotFound) {
				SendJsonResponse(w, http.StatusNotFound, nil, "Instance not found")
				return
			}

			log.Error().Err(res.Error).Msgf("Error finding instance")
			SendJsonResponse(w, http.StatusInternalServerError, nil, res.Error.Error())
			return
		}

		msg := a.ResourceNames.Singular + " Detail"

		// 5. Send Success Response
		SendJsonResponse(w, http.StatusOK, instance, msg)
	}
}
