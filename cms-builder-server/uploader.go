package builder

import (
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"
)

type File struct {
	*SystemData
	Name     string `json:"name"`
	Path     string `json:"path"` // relative path
	Url      string `json:"url"`  // absolute path
	Size     int64  `json:"size"`
	MimeType string `json:"mimeType"`
}

type UploaderConfig struct {
	MaxSize            int64
	SupportedMimeTypes []string
	Folder             string
}

var CreateStoredFilesHandler ApiFunction = func(a *App, db *Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		err := ValidateRequestMethod(r, http.MethodPost)
		if err != nil {
			SendJsonResponse(w, http.StatusMethodNotAllowed, nil, err.Error())
			return
		}

		params := FormatRequestParameters(r, a.Admin.Builder)
		isAllowed := a.Permissions.HasPermission(params.Roles, OperationCreate)
		if !isAllowed {
			SendJsonResponse(w, http.StatusForbidden, nil, "User is not allowed to create this resource")
			return
		}

		cfg := &UploaderConfig{
			MaxSize:            config.GetInt64(config.GetString(EnvKeys.UploaderMaxSize)),
			SupportedMimeTypes: config.GetStringSlice(EnvKeys.UploaderSupportedMime),
			Folder:             config.GetString(config.GetString(EnvKeys.UploaderFolder)),
		}

		// Store the file
		err = r.ParseMultipartForm(cfg.MaxSize)
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

		// Check if the file type is supported
		contentType := header.Header.Get("Content-Type")

		validContentType, err := ValidateContentType(contentType, cfg.SupportedMimeTypes)
		if err != nil {
			SendJsonResponse(w, http.StatusBadRequest, nil, err.Error())
			return
		}

		if !validContentType {
			SendJsonResponse(w, http.StatusBadRequest, nil, "Invalid file type")
			return
		}

		store := a.Admin.Builder.Store

		fileName := header.Filename
		fileData, err := store.StoreFile(cfg, fileName, file)
		if err != nil {
			handleUploadError(store, fileData, w, err)
			return
		}

		fileInfo, err := store.GetFileInfo(&fileData)
		if err != nil {
			log.Error().Err(err).Msg("Error getting file info")
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}

		fileData.SystemData = &SystemData{
			CreatedByID: params.User.ID,
			UpdatedByID: params.User.ID,
		}
		fileData.Name = header.Filename
		fileData.Size = fileInfo.Size
		fileData.MimeType = fileInfo.ContentType

		// Run validations
		validationErrors := a.Validate(fileData)
		if len(validationErrors.Errors) > 0 {
			SendJsonResponse(w, http.StatusBadRequest, validationErrors, "Validation failed")
			return
		}

		res := db.Create(fileData, params.User)
		if res.Error != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, res.Error.Error())
			return
		}

		SendJsonResponse(w, http.StatusCreated, &fileData, a.Name()+" created")
	}
}

var DeleteStoredFilesHandler ApiFunction = func(a *App, db *Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		err := ValidateRequestMethod(r, http.MethodDelete)
		if err != nil {
			SendJsonResponse(w, http.StatusMethodNotAllowed, nil, err.Error())
			return
		}

		params := FormatRequestParameters(r, a.Admin.Builder)
		isAllowed := a.Permissions.HasPermission(params.Roles, OperationDelete)
		if !isAllowed {
			SendJsonResponse(w, http.StatusForbidden, nil, "User is not allowed to delete this resource")
			return
		}

		instanceId := GetUrlParam("id", r)

		instance, err := GetInstanceIfAuthorized(a.Model, a.SkipUserBinding, instanceId, db, &params)
		if err != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}

		store := a.Admin.Builder.Store

		fileData := instance.(*File)

		err = store.DeleteFile(*fileData)
		if err != nil {
			handleUploadError(store, *fileData, w, err)
			return
		}

		res := db.Delete(fileData, params.User)
		if res.Error != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, res.Error.Error())
			return
		}

		// Send a 204 No Content response
		SendJsonResponse(w, http.StatusOK, nil, a.Name()+" deleted")

	}
}

var UpdateStoredFilesHandler ApiFunction = func(a *App, db *Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := ValidateRequestMethod(r, http.MethodPut)
		if err != nil {
			SendJsonResponse(w, http.StatusMethodNotAllowed, nil, err.Error())
			return
		}

		SendJsonResponse(w, http.StatusMethodNotAllowed, nil, "You cannot update a file. You may delete and create a new one.")
	}
}

func (b *Builder) DownloadStoredFileHandler() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		err := ValidateRequestMethod(r, http.MethodGet)
		if err != nil {
			SendJsonResponse(w, http.StatusMethodNotAllowed, err, err.Error())
			return
		}

		app, err := b.Admin.GetApp("file")
		if err != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}

		params := FormatRequestParameters(r, b)
		isAllowed := app.Permissions.HasPermission(params.Roles, OperationRead)
		if !isAllowed {
			SendJsonResponse(w, http.StatusForbidden, nil, "User is not allowed to read this resource")
			return
		}

		// Create a new instance of the model
		instanceId := GetUrlParam("id", r)

		instance, err := GetInstanceIfAuthorized(app.Model, app.SkipUserBinding, instanceId, b.DB, &params)
		if err != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}

		file := instance.(*File)
		if file == (*File)(nil) {
			SendJsonResponse(w, http.StatusNotFound, nil, "File not found")
			return
		}

		bytes, err := b.Store.ReadFile(file)
		if err != nil {
			log.Error().Err(err).Msg("Error reading file")
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}

		w.Write(bytes)
	}
}

// getMimeTypeAndExtension takes a mime string and returns the mime type and extension
// split by the "/" character. For example, "audio/wav" would return "audio" and "wav".
func getMimeTypeAndExtension(mime string) (string, string) {
	parts := strings.Split(mime, "/")
	if len(parts) == 0 {
		return "", "" // Or handle this as an error if you prefer
	}
	mimeType := parts[0]
	extension := ""
	if len(parts) > 1 {
		extension = parts[len(parts)-1] // Get the last part as the extension
	}
	return mimeType, extension
}

// ValidateContentType takes a content type and a list of supported mime types
// and returns true if the content type is supported, and false otherwise.
// It also returns an error if the content type is not supported.
//
// The supported mime types can be in different formats, such as:
// - "*": Supports all mime types.
// - "*/*": Supports all mime types with a specific mime type.
// - "audio/*": Supports all mime types with a specific mime type and extension.
// - "audio/wav": Supports a specific mime type and extension.
func ValidateContentType(contentType string, supportedMimeTypes []string) (bool, error) {
	inMimeType, inExtension := getMimeTypeAndExtension(contentType)

	for _, supportedItem := range supportedMimeTypes {
		// "*"
		if supportedItem == "*" {
			return true, nil
		}
		supportedMimeType, supportedExtension := getMimeTypeAndExtension(supportedItem)
		// "*/*"
		if supportedMimeType == "*" {
			return true, nil
		}
		// "audio/*"
		if supportedExtension == "*" && inMimeType == supportedMimeType {
			return true, nil
		}
		// "audio/wav"
		if supportedExtension == inExtension && supportedMimeType == inMimeType {
			return true, nil
		}
	}

	return false, nil
}

// handleUploadError takes a file path and a writer, and an error.
// It logs the error, deletes the file from disk, and writes a JSON response
// with the error message to the writer.
func handleUploadError(store Store, fileData File, w http.ResponseWriter, err error) {
	// Log the error at the error level
	log.Error().Err(err).Msgf("Error uploading file: %s. Rolling back...", fileData.Name)

	// Attempt to delete the file from disk
	store.DeleteFile(fileData)

	// Write a JSON response with the error message to the writer
	// at the internal server error (500) status code.
	SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
}

// randomizeFileName takes a file name and returns a new file name that is
// randomized using the current timestamp. The file name is also sanitized
// to replace spaces, forward slashes, and backslashes with underscores.
func randomizeFileName(fileName string) string {

	name, extension := path.Split(fileName)

	// Replace spaces with underscores
	name = strings.ReplaceAll(name, " ", "_")

	// Replace forward slashes with underscores
	name = strings.ReplaceAll(name, "/", "_")

	// Replace backslashes with underscores
	name = strings.ReplaceAll(name, "\\", "_")

	// Add the current timestamp to the file name
	now := strconv.FormatInt(time.Now().UnixNano(), 10)
	name = now + "_" + name

	return name + "." + extension
}
