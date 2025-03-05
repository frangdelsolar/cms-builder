package file

import (
	"net/http"

	authConstants "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/constants"
	authModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/models"
	authUtils "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/utils"
	dbQueries "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/queries"
	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"
	fileModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/file/models"
	loggerTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger/types"
	rmTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager/types"
	svrUtils "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server/utils"
	storeTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/store/types"
	utilsPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/utils"
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

			fileName := header.Filename
			fileData, err := st.StoreFile(fileName, file, header, log)
			if err != nil {
				handleUploadError(st, fileData, w, err, log)
				return
			}

			fileData.SystemData = &authModels.SystemData{
				CreatedByID: user.ID,
				UpdatedByID: user.ID,
			}

			// Run validations
			validationErrors := a.Validate(fileData, log)
			if len(validationErrors.Errors) > 0 {
				svrUtils.SendJsonResponse(w, http.StatusBadRequest, validationErrors, "Validation failed")
				return
			}

			err = dbQueries.Create(r.Context(), log, db, fileData, user, requestId)
			if err != nil {
				svrUtils.SendJsonResponse(w, http.StatusInternalServerError, nil, "Error creating "+a.ResourceNames.Singular)
				return
			}

			// Generate url and update
			fdCopy := fileData
			differences := utilsPkg.CompareInterfaces(&fdCopy, fileData)

			fileData.Url = apiBaseUrl + "/private/api/files/" + fileData.StringID() + "/download"

			err = dbQueries.Update(r.Context(), log, db, fileData, user, differences, requestId)
			if err != nil {
				svrUtils.SendJsonResponse(w, http.StatusInternalServerError, nil, "Error updating "+a.ResourceNames.Singular)
				return
			}

			svrUtils.SendJsonResponse(w, http.StatusCreated, &fileData, a.ResourceNames.Singular+" created")
		}
	}
}

func handleUploadError(store storeTypes.Store, fileData *fileModels.File, w http.ResponseWriter, err error, log *loggerTypes.Logger) {
	log.Error().Err(err).Msgf("Error uploading file: %s. Rolling back...", fileData.Name)

	// Attempt to delete the file from disk
	store.DeleteFile(fileData, log)

	// Write a JSON response with the error message to the writer
	// at the internal server error (500) status code.
	svrUtils.SendJsonResponse(w, http.StatusBadRequest, nil, err.Error())
}
