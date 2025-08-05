package auth

import (
	"encoding/json"
	"fmt"
	"net/http"

	authConstants "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/constants"
	authUtils "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/utils"
	dbQueries "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/queries"
	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"
	rmTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager/types"
	svrUtils "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server/utils"
	utilsPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/utils"
)

var UserUpdateHandler rmTypes.ApiFunction = func(a *rmTypes.Resource, db *dbTypes.DatabaseConnection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestCtx := svrUtils.GetRequestContext(r)
		log := requestCtx.Logger
		user := requestCtx.User
		requestId := requestCtx.RequestId
		isAdmin := user.HasRole(authConstants.AdminRole)

		// 1. Validate Request Method
		err := svrUtils.ValidateRequestMethod(r, http.MethodPut)
		if err != nil {
			svrUtils.SendJsonResponse(w, http.StatusMethodNotAllowed, nil, err.Error())
			return
		}

		// 2. Check Permissions
		if !authUtils.UserIsAllowed(a.Permissions, user.GetRoles(), authConstants.OperationRead, a.ResourceNames.Singular, log) {
			svrUtils.SendJsonResponse(w, http.StatusForbidden, nil, "User is not allowed to access this resource")
			return
		}

		if !authUtils.UserIsAllowed(a.Permissions, user.GetRoles(), authConstants.OperationUpdate, a.ResourceNames.Singular, log) {
			svrUtils.SendJsonResponse(w, http.StatusForbidden, nil, "User is not allowed to update this resource")
			return
		}

		if !isAdmin {
			if user.StringID() != svrUtils.GetUrlParam("id", r) {
				svrUtils.SendJsonResponse(w, http.StatusNotFound, nil, "Instance not found")
				return
			}
		}

		filters := map[string]interface{}{
			"id": svrUtils.GetUrlParam("id", r),
		}

		instance := a.GetOne()
		err = dbQueries.FindOne(r.Context(), log, db, &instance, filters, []string{})
		if err != nil {
			log.Error().Err(err).Msgf("Instance not found")
			svrUtils.SendJsonResponse(w, http.StatusNotFound, nil, "Instance not found")
			return
		}

		previousState := a.GetOne()
		_ = dbQueries.FindOne(r.Context(), log, db, &previousState, filters, []string{})

		// 5. Format Request Body and Filter Keys
		body, err := svrUtils.FormatRequestBody(r, filterKeys)
		if err != nil {
			fmt.Printf("Error formatting request body: %v\n", err)
			log.Error().Err(err).Msg("Error formatting request body")
			svrUtils.SendJsonResponse(w, http.StatusBadRequest, nil, "Invalid request body")
			return
		}

		// 7. Marshal Body to JSON
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			fmt.Printf("Error marshalling request body: %v\n", err)
			log.Error().Err(err).Msg("Error marshalling request body")
			svrUtils.SendJsonResponse(w, http.StatusInternalServerError, nil, "Invalid request body")
			return
		}

		// 8. Unmarshal Body into Instance
		err = json.Unmarshal(bodyBytes, &instance)
		if err != nil {
			fmt.Printf("Error unmarshalling request body: %v\n", err)
			log.Error().Err(err).Msg("Error unmarshalling request body")
			svrUtils.SendJsonResponse(w, http.StatusInternalServerError, nil, "Invalid request body")
			return
		}

		// 9. Run Validations
		validationErrors := a.Validate(instance, log)
		if len(validationErrors.Errors) > 0 {
			svrUtils.SendJsonResponse(w, http.StatusBadRequest, validationErrors, "Validation failed")
			return
		}

		// 10. Find differences with existing instance
		differences := utilsPkg.CompareInterfaces(previousState, instance)
		if diffMap, ok := differences.(map[string]interface{}); ok && len(diffMap) == 0 {
			svrUtils.SendJsonResponse(w, http.StatusOK, instance, a.ResourceNames.Singular+" is up to date")
			return
		}

		// 11. Create Instance in Database
		err = dbQueries.Update(r.Context(), log, db, instance, user, differences, requestId)
		if err != nil {
			svrUtils.SendJsonResponse(w, http.StatusInternalServerError, nil, "Error updating resource")
			return
		}

		msg := a.ResourceNames.Singular + " has been updated"

		// 12. Send Success Response
		svrUtils.SendJsonResponse(w, http.StatusOK, instance, msg)
	}
}
