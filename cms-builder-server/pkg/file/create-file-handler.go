package file

import (
	"net/http"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/queries"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	manager "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/store"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/utils"
)

func CreateStoredFilesHandler(db *database.Database, st store.Store, apiBaseUrl string) manager.ApiFunction {

	return func(a *manager.Resource, db *database.Database) http.HandlerFunc {
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

			// 3. Check Permissions
			if !UserIsAllowed(a.Permissions, user.GetRoles(), OperationCreate, a.ResourceNames.Singular, log) {
				SendJsonResponse(w, http.StatusForbidden, nil, "User is not allowed to read this resource")
				return
			}

			// Store the file
			storeConfig := st.GetConfig()
			err = r.ParseMultipartForm(storeConfig.MaxSize)
			if err != nil {
				log.Error().Err(err).Msg("Error parsing multipart form data")
				SendJsonResponse(w, http.StatusBadRequest, nil, err.Error())
				return
			}

			// Get the file from the form
			file, header, err := r.FormFile("file")
			if err != nil {
				SendJsonResponse(w, http.StatusBadRequest, nil, err.Error())
				return
			}
			defer file.Close()

			fileName := header.Filename
			fileData, err := st.StoreFile(fileName, file, header, log)
			if err != nil {
				handleUploadError(st, fileData, w, err, log)
				return
			}

			fileData.SystemData = &models.SystemData{
				CreatedByID: user.ID,
				UpdatedByID: user.ID,
			}

			// Run validations
			validationErrors := a.Validate(fileData, log)
			if len(validationErrors.Errors) > 0 {
				SendJsonResponse(w, http.StatusBadRequest, validationErrors, "Validation failed")
				return
			}

			res := queries.Create(db, fileData, user, requestId)
			if res.Error != nil {
				SendJsonResponse(w, http.StatusInternalServerError, nil, res.Error.Error())
				return
			}

			// Generate url and update
			fdCopy := fileData
			differences := utils.CompareInterfaces(&fdCopy, fileData)

			fileData.Url = apiBaseUrl + "/private/api/files/" + fileData.StringID() + "/download"

			res = queries.Update(db, fileData, user, differences, requestId)
			if res.Error != nil {
				SendJsonResponse(w, http.StatusInternalServerError, nil, res.Error.Error())
				return
			}

			SendJsonResponse(w, http.StatusCreated, &fileData, a.ResourceNames.Singular+" created")
		}
	}
}

func handleUploadError(store store.Store, fileData *models.File, w http.ResponseWriter, err error, log *logger.Logger) {
	log.Error().Err(err).Msgf("Error uploading file: %s. Rolling back...", fileData.Name)

	// Attempt to delete the file from disk
	store.DeleteFile(fileData, log)

	// Write a JSON response with the error message to the writer
	// at the internal server error (500) status code.
	SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
}
