package auth

import (
	"encoding/json"
	"fmt"
	"net/http"

	authConstants "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/constants"
	authModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/models"
	authTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/types"
	authUtils "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/utils"
	cliPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/clients"
	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"
	svrUtils "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server/utils"
)

func RegisterVisitorController(firebase *cliPkg.FirebaseManager, db *dbTypes.DatabaseConnection, getSystemUser func() *authModels.User) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		requestCtx := svrUtils.GetRequestContext(r)
		log := requestCtx.Logger
		requestId := requestCtx.RequestId

		// 1. Validate Request Method
		err := svrUtils.ValidateRequestMethod(r, http.MethodPost)
		if err != nil {
			svrUtils.SendJsonResponse(w, http.StatusMethodNotAllowed, nil, err.Error())
			return
		}

		// 2. Format Request Body
		body, err := svrUtils.FormatRequestBody(r, map[string]bool{
			authConstants.RolesParamKey: true,
		})
		if err != nil {
			svrUtils.SendJsonResponse(w, http.StatusBadRequest, nil, err.Error())
			return
		}

		// 3. Marshal Body to JSON
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			log.Error().Err(err).Msg("Error marshalling request body")
			svrUtils.SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}

		// 4. Unmarshal Body into Instance
		input := authTypes.RegisterUserInput{}
		err = json.Unmarshal(bodyBytes, &input)
		if err != nil {
			svrUtils.SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}

		input.Roles = []authTypes.Role{
			authConstants.VisitorRole,
		}
		input.RegisterFirebase = true

		systemUser := getSystemUser()
		if systemUser == nil {
			log.Error().Msg("System user not found")
			svrUtils.SendJsonResponse(w, http.StatusInternalServerError, nil, "System user not found")
			return
		}

		// 5. Create User
		user, err := authUtils.CreateUserWithRole(input, firebase, db, systemUser, requestId, log)
		if err != nil {
			msg := fmt.Sprintf("Error creating user: %s", err.Error())
			svrUtils.SendJsonResponse(w, http.StatusInternalServerError, nil, msg)
			return
		}

		svrUtils.SendJsonResponse(w, http.StatusOK, user, "User registered successfully")
	}
}
