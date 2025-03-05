package auth

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	dbQueries "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/queries"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	mgr "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/utils"
)

var UserUpdateHandler mgr.ApiFunction = func(a *mgr.Resource, db *database.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestCtx := GetRequestContext(r)
		log := requestCtx.Logger
		user := requestCtx.User
		requestId := requestCtx.RequestId
		isAdmin := user.HasRole(models.AdminRole)

		// 1. Validate Request Method
		err := ValidateRequestMethod(r, http.MethodPut)
		if err != nil {
			SendJsonResponse(w, http.StatusMethodNotAllowed, nil, err.Error())
			return
		}

		// 2. Check Permissions
		if !UserIsAllowed(a.Permissions, user.GetRoles(), OperationRead, a.ResourceNames.Singular, log) {
			SendJsonResponse(w, http.StatusForbidden, nil, "User is not allowed to access this resource")
			return
		}

		if !UserIsAllowed(a.Permissions, user.GetRoles(), OperationUpdate, a.ResourceNames.Singular, log) {
			SendJsonResponse(w, http.StatusForbidden, nil, "User is not allowed to update this resource")
			return
		}

		if !isAdmin {
			if user.StringID() != GetUrlParam("id", r) {
				SendJsonResponse(w, http.StatusNotFound, nil, "Instance not found")
				return
			}
		}

		filters := map[string]interface{}{
			"id": GetUrlParam("id", r),
		}

		instance := a.GetOne()
		err = dbQueries.FindOne(r.Context(), log, db, &instance, filters)
		if err != nil {
			log.Error().Err(err).Msgf("Instance not found")
			SendJsonResponse(w, http.StatusNotFound, nil, "Instance not found")
			return
		}

		previousState := a.GetOne()
		_ = dbQueries.FindOne(r.Context(), log, db, &previousState, filters)

		// 5. Format Request Body and Filter Keys
		body, err := FormatRequestBody(r, filterKeys)
		if err != nil {
			fmt.Printf("Error formatting request body: %v\n", err)
			log.Error().Err(err).Msg("Error formatting request body")
			SendJsonResponse(w, http.StatusBadRequest, nil, "Invalid request body")
			return
		}

		// 7. Marshal Body to JSON
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			fmt.Printf("Error marshalling request body: %v\n", err)
			log.Error().Err(err).Msg("Error marshalling request body")
			SendJsonResponse(w, http.StatusInternalServerError, nil, "Invalid request body")
			return
		}

		// 8. Unmarshal Body into Instance
		err = json.Unmarshal(bodyBytes, &instance)
		if err != nil {
			fmt.Printf("Error unmarshalling request body: %v\n", err)
			log.Error().Err(err).Msg("Error unmarshalling request body")
			SendJsonResponse(w, http.StatusInternalServerError, nil, "Invalid request body")
			return
		}

		// 9. Run Validations
		validationErrors := a.Validate(instance, log)
		if len(validationErrors.Errors) > 0 {
			SendJsonResponse(w, http.StatusBadRequest, validationErrors, "Validation failed")
			return
		}

		// 10. Find differences with existing instance
		differences := utils.CompareInterfaces(previousState, instance)
		if diffMap, ok := differences.(map[string]interface{}); ok && len(diffMap) == 0 {
			SendJsonResponse(w, http.StatusOK, instance, a.ResourceNames.Singular+" is up to date")
			return
		}

		// 11. Create Instance in Database
		err = dbQueries.Update(r.Context(), log, db, instance, user, differences, requestId)
		if err != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, "Error updating resource")
			return
		}

		msg := a.ResourceNames.Singular + " has been updated"

		// 12. Send Success Response
		SendJsonResponse(w, http.StatusOK, instance, msg)
	}
}
