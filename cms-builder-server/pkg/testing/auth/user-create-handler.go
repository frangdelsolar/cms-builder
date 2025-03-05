package auth

import (
	"encoding/json"
	"net/http"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/queries"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"

	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"
	mgr "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"
)

var filterKeys = map[string]bool{
	"ID": true,
}

// DefaultCreateHandler handles the creation of a new resource.
var UserCreateHandler mgr.ApiFunction = func(a *mgr.Resource, db *dbTypes.DatabaseConnection) http.HandlerFunc {
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
		if !UserIsAllowed(a.Permissions, user.GetRoles(), OperationCreate, a.ResourceNames.Singular, log) { // corrected to OperationCreate
			SendJsonResponse(w, http.StatusForbidden, nil, "User is not allowed to create this resource")
			return
		}

		// 3. Format Request Body and Filter Keys
		body, err := FormatRequestBody(r, filterKeys)
		if err != nil {
			SendJsonResponse(w, http.StatusBadRequest, nil, "Invalid request body")
			return
		}

		// 5. Marshal Body to JSON
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			log.Error().Err(err).Msg("Error marshalling request body")
			SendJsonResponse(w, http.StatusInternalServerError, nil, "Invalid request body")
			return
		}

		// 6. Unmarshal Body into Instance
		instance := a.GetOne()
		err = json.Unmarshal(bodyBytes, &instance)
		if err != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, "Invalid request body")
			return
		}

		// 7. Run Validations
		validationErrors := a.Validate(instance, log)
		if len(validationErrors.Errors) > 0 {
			SendJsonResponse(w, http.StatusBadRequest, validationErrors, "Validation failed")
			return
		}

		// 8. Create Instance in Database
		err = queries.Create(r.Context(), log, db, instance, user, requestId)
		if err != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, "Error creating resource")
			return
		}

		// 9. Generate Success Message
		msg := a.ResourceNames.Singular + " has been created"

		// 10. Send Success Response
		SendJsonResponse(w, http.StatusCreated, &instance, msg)
	}
}
