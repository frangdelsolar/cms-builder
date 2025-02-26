package resourcemanager

import (
	"encoding/json"
	"net/http"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/queries"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/utils"
)

var DefaultUpdateHandler ApiFunction = func(a *Resource, db *database.Database) http.HandlerFunc {
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
			SendJsonResponse(w, http.StatusForbidden, nil, "User is not allowed to delete this resource")
			return
		}

		// 3. Construct Query (User Binding)
		query := ""
		if !(a.SkipUserBinding || isAdmin) {
			query = "created_by_id = ?"
		}

		// 4. Retrieve Instance ID and Fetch Instance
		instanceId := GetUrlParam("id", r)
		instance := a.GetOne()
		res := queries.FindOne(db, instance, instanceId, query, user.StringID())

		if res.Error != nil {
			log.Error().Err(res.Error).Msgf("Error finding instance")
			SendJsonResponse(w, http.StatusInternalServerError, nil, res.Error.Error())
			return
		}

		if instance == nil {
			log.Error().Msgf("Instance not found")
			SendJsonResponse(w, http.StatusNotFound, nil, "Instance not found")
			return
		}

		previousState := a.GetOne()
		_ = queries.FindOne(db, previousState, instanceId, query, user.StringID())

		// 5. Format Request Body and Filter Keys
		body, err := FormatRequestBody(r, filterKeys)
		if err != nil {
			SendJsonResponse(w, http.StatusBadRequest, nil, err.Error())
			return
		}

		// 6. Add User Information
		body["UpdatedByID"] = user.StringID()

		// 7. Marshal Body to JSON
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			log.Error().Err(err).Msg("Error marshalling request body")
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}

		// 8. Unmarshal Body into Instance
		err = json.Unmarshal(bodyBytes, &instance)
		if err != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
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
		res = queries.Update(db, instance, user, differences, requestId)
		if res.Error != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, res.Error.Error())
			return
		}

		msg := a.ResourceNames.Singular + " has been updated"

		// 12. Send Success Response
		SendJsonResponse(w, http.StatusOK, instance, msg)
	}
}
