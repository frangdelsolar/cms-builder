package file

import (
	"net/http"

	authConstants "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/constants"
	authUtils "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/utils"
	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"
	rmTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager/types"
	svrUtils "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server/utils"
	storeTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/store/types"

	fileUtils "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/file/utils"
)

func CreateStoredFilesHandler(db *dbTypes.DatabaseConnection, st storeTypes.Store, apiBaseUrl string) rmTypes.ApiFunction {

	return func(a *rmTypes.Resource, db *dbTypes.DatabaseConnection) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {

			requestCtx := svrUtils.GetRequestContext(r)
			log := requestCtx.Logger
			user := requestCtx.User
			requestId := requestCtx.RequestId

			// 1. Validate Request Method
			err := svrUtils.ValidateRequestMethod(r, http.MethodPost)
			if err != nil {
				svrUtils.SendJsonResponse(w, http.StatusMethodNotAllowed, nil, err.Error())
				return
			}

			// 3. Check Permissions
			if !authUtils.UserIsAllowed(a.Permissions, user.GetRoles(), authConstants.OperationCreate, a.ResourceNames.Singular, log) {
				svrUtils.SendJsonResponse(w, http.StatusForbidden, nil, "User is not allowed to read this resource")
				return
			}

			// Store the file
			storeConfig := st.GetConfig()
			err = r.ParseMultipartForm(storeConfig.MaxSize)
			if err != nil {
				log.Error().Err(err).Msg("Error parsing multipart form data")
				svrUtils.SendJsonResponse(w, http.StatusBadRequest, nil, err.Error())
				return
			}

			// Get the file from the form
			file, header, err := r.FormFile("file")
			if err != nil {
				svrUtils.SendJsonResponse(w, http.StatusBadRequest, nil, err.Error())
				return
			}
			defer file.Close()

			storedFile, err := fileUtils.StoreUploadedFile(
				r.Context(),
				log,
				db,
				st,
				apiBaseUrl,
				a, // resource
				file,
				header,
				user,
				requestId,
			)

			if err != nil {
				svrUtils.SendJsonResponse(w, http.StatusBadRequest, nil, err.Error())
				return
			}

			svrUtils.SendJsonResponse(w, http.StatusCreated, storedFile, a.ResourceNames.Singular+" created")
		}
	}
}
