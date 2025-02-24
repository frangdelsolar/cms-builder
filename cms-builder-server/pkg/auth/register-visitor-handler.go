package auth

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/clients"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
)

func RegisterVisitorController(firebase *clients.FirebaseManager, db *database.Database, systemUser *models.User) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		requestCtx := GetRequestContext(r)
		log := requestCtx.Logger
		requestId := requestCtx.RequestId

		// 1. Validate Request Method
		err := ValidateRequestMethod(r, http.MethodPost)
		if err != nil {
			SendJsonResponse(w, http.StatusMethodNotAllowed, nil, err.Error())
			return
		}

		// 2. Format Request Body
		body, err := FormatRequestBody(r, map[string]bool{
			RolesParamKey.S(): true,
		})
		if err != nil {
			SendJsonResponse(w, http.StatusBadRequest, nil, err.Error())
			return
		}

		// 3. Marshal Body to JSON
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			log.Error().Err(err).Msg("Error marshalling request body")
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}

		// 4. Unmarshal Body into Instance
		input := models.RegisterUserInput{}
		err = json.Unmarshal(bodyBytes, &input)
		if err != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}

		input.Roles = []models.Role{
			models.VisitorRole,
		}

		// 5. Create User
		user, err := CreateUserWithRole(input, firebase, db, systemUser, requestId)
		if err != nil {
			msg := fmt.Sprintf("Error creating user: %s", err.Error())
			SendJsonResponse(w, http.StatusInternalServerError, nil, msg)
			return
		}

		SendJsonResponse(w, http.StatusOK, user, "User registered successfully")
	}
}
