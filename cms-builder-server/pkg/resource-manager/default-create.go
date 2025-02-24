package resourcemanager

// import (
// 	"net/http"

// 	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
// 	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
// 	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/queries"
// 	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
// )
// var DefaultCreateHandler ApiFunction = func(a *Resource, db *database.Database) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {

// 		err := ValidateRequestMethod(r, http.MethodPost)
// 		if err != nil {
// 			SendJsonResponse(w, http.StatusMethodNotAllowed, nil, err.Error())
// 			return
// 		}

// 		requestCtx := GetRequestContext(r)
// 		log := requestCtx.Logger

// 		params := FormatRequestParameters(r, a.Admin.Builder)
// 		isAllowed := a.Permissions.HasPermission(params.Roles, OperationCreate)
// 		if !isAllowed {
// 			SendJsonResponse(w, http.StatusForbidden, nil, "User is not allowed to create this resource")
// 			return
// 		}

// 		// Create a new instance of the model and parse the request body
// 		body, err := FormatRequestBody(r, filterKeys)
// 		if err != nil {
// 			SendJsonResponse(w, http.StatusBadRequest, nil, err.Error())
// 			return
// 		}

// 		body["CreatedByID"] = params.User.ID
// 		body["UpdatedByID"] = params.User.ID

// 		bodyBytes, err := json.Marshal(body)
// 		if err != nil {
// 			log.Error().Err(err).Msg("Error marshalling request body")
// 			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
// 			return
// 		}

// 		instance := CreateInstanceForUndeterminedType(a.Model)
// 		err = json.Unmarshal(bodyBytes, &instance)
// 		if err != nil {
// 			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
// 			return
// 		}

// 		// Run validations
// 		validationErrors := a.Validate(instance)
// 		if len(validationErrors.Errors) > 0 {
// 			SendJsonResponse(w, http.StatusBadRequest, validationErrors, "Validation failed")
// 			return
// 		}

// 		res := db.Create(instance, params.User, params.RequestId)
// 		if res.Error != nil {
// 			SendJsonResponse(w, http.StatusInternalServerError, nil, res.Error.Error())
// 			return
// 		}

// 		SendJsonResponse(w, http.StatusCreated, &instance, a.Name()+" created")
// 	}
// }
