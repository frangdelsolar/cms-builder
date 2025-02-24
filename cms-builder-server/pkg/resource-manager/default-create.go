package resourcemanager

import (
	"encoding/json"
	"net/http"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/queries"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
)

// filterKeys defines the keys to be filtered out from the request body.
var filterKeys = map[string]bool{
	"id":            true,
	"createdBy":     true,
	"created_by":    true,
	"createdById":   true,
	"created_by_id": true,
	"updatedBy":     true,
	"updated_by":    true,
	"updatedById":   true,
	"updated_by_id": true,
	"deletedBy":     true,
	"deleted_by":    true,
	"deletedById":   true,
	"deleted_by_id": true,
	"createdAt":     true,
	"created_at":    true,
	"updatedAt":     true,
	"updated_at":    true,
	"deletedAt":     true,
	"deleted_at":    true,
}

// DefaultCreateHandler handles the creation of a new resource.
var DefaultCreateHandler ApiFunction = func(a *Resource, db *database.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestCtx := GetRequestContext(r)
		log := requestCtx.Logger
		user := requestCtx.User
		requestId := requestCtx.RequestId

		// 1. Validate Request Method
		err := ValidateRequestMethod(r, http.MethodPost)
		if err != nil {
			SendJsonResponse(w, http.StatusMethodNotAllowed, nil, err.Error())
			return
		}

		// 2. Check Permissions
		if !UserIsAllowed(a.Permissions, user.GetRoles(), OperationCreate) { // corrected to OperationCreate
			SendJsonResponse(w, http.StatusForbidden, nil, "User is not allowed to create this resource")
			return
		}

		// 3. Format Request Body and Filter Keys
		body, err := FormatRequestBody(r, filterKeys)
		if err != nil {
			SendJsonResponse(w, http.StatusBadRequest, nil, err.Error())
			return
		}

		// 4. Add User Information
		body["CreatedByID"] = user.StringID()
		body["UpdatedByID"] = user.StringID()

		// 5. Marshal Body to JSON
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			log.Error().Err(err).Msg("Error marshalling request body")
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}

		// 6. Unmarshal Body into Instance
		instance := a.GetOne()
		err = json.Unmarshal(bodyBytes, &instance)
		if err != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}

		// 7. Run Validations
		validationErrors := a.Validate(instance)
		if len(validationErrors.Errors) > 0 {
			SendJsonResponse(w, http.StatusBadRequest, validationErrors, "Validation failed")
			return
		}

		// 8. Create Instance in Database
		res := queries.Create(db, instance, user, requestId)
		if res.Error != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, res.Error.Error())
			return
		}

		// 9. Generate Success Message
		appName, err := a.GetName()
		if err != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}
		msg := appName + " created"

		// 10. Send Success Response
		SendJsonResponse(w, http.StatusCreated, &instance, msg)
	}
}
