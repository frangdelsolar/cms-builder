package resourcemanager

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/queries"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
	"gorm.io/gorm"
)

// DefaultDeleteHandler handles the deletion of a resource.
var DefaultDeleteHandler ApiFunction = func(a *Resource, db *database.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestCtx := GetRequestContext(r)
		log := requestCtx.Logger
		user := requestCtx.User
		requestId := requestCtx.RequestId
		isAdmin := user.HasRole(models.AdminRole)

		fmt.Println("Deleting Instance")

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
		query := "id = ?"
		if !(a.SkipUserBinding || isAdmin) {
			query += " AND created_by_id = ?"
		}

		// 4. Retrieve Instance ID and Fetch Instance
		instanceId := GetUrlParam("id", r)
		instance := a.GetOne()

		res := queries.FindOne(db, instance, query, instanceId, user.StringID())
		if res.Error != nil {
			if errors.Is(res.Error, gorm.ErrRecordNotFound) {
				SendJsonResponse(w, http.StatusNotFound, nil, "Instance not found")
				return
			}
			log.Error().Err(res.Error).Msgf("Error finding instance")
			SendJsonResponse(w, http.StatusInternalServerError, nil, "Error finding resource")
			return
		}

		// 5. Delete Instance
		res = queries.Delete(db, instance, user, requestId)
		if res.Error != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, "Error deleting resource")
			return
		}

		// 6. Generate Success Message
		msg := a.ResourceNames.Singular + " has been deleted"

		// 7. Send Success Response
		SendJsonResponse(w, http.StatusOK, nil, msg)
	}
}
