package builder

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type FileData struct {
	Name string `json:"fileName"`
	Path string `json:"filePath"` // relative path
	Url  string `json:"url"`      // absolute path
}

type Upload struct {
	*SystemData
	*FileData
}

type UploaderConfig struct {
	MaxSize            int64
	SupportedMimeTypes []string
	Folder             string
	StaticPath         string
}

// getUploadPostHandler returns a handler function that responds to POST requests
// on the uploads endpoint, e.g. /api/uploads.
//
// The handler function will save the uploaded file to disk and store the file
// information in the database. It will also handle errors and return a 400
// error if the request body is not valid JSON, or a 500 error if there is an
// error storing the file or saving the file information to the database.
func (b *Builder) GetFilePostHandler(cfg *UploaderConfig) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Validate method
		err := ValidateRequestMethod(r, http.MethodPost)
		if err != nil {
			SendJsonResponse(w, http.StatusMethodNotAllowed, nil, err.Error())
			return
		}

		uploadApp, err := b.Admin.GetApp("upload")
		if err != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}

		params := FormatRequestParameters(r, b)
		isAllowed := uploadApp.Permissions.HasPermission(params.Roles, OperationCreate)
		if !isAllowed {
			SendJsonResponse(w, http.StatusForbidden, nil, "User is not allowed to create this resource")
			return
		}

		// Parse the multipart form data
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

		fileName := header.Filename
		fileData, err := b.Store.StoreFile(cfg, fileName, file)
		if err != nil {
			handleUploadError(b.Store, fileData, w, err)
			return
		}

		uploadRequestBody := map[string]interface{}{
			"fileName": fileData.Name,
			"filePath": fileData.Path,
			"url":      fileData.Url,
		}

		uploadData, err := json.Marshal(uploadRequestBody)
		if err != nil {
			handleUploadError(b.Store, fileData, w, err)
			return
		}

		request := &http.Request{
			Method: http.MethodPost,
			Header: r.Header,
			Body:   io.NopCloser(bytes.NewBuffer(uploadData)),
		}

		// This will send the response to the client
		uploadApp.ApiCreate(b.DB)(w, request)
	}
}

// getUploadDeleteHandler returns a handler function that responds to DELETE
// requests on the delete file endpoint, e.g. /file/{id}/delete.
//
// It will delete the file from disk and remove the record from the database.
// If the file is not found, it will return a 404 Not Found response. If the
// database record is not found, it will return a 500 Internal Server Error
// response. If the file is deleted successfully, it will return a 200 OK
// response with a message saying "File deleted successfully".
func (b *Builder) GetFileDeleteHandler(cfg *UploaderConfig) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		err := ValidateRequestMethod(r, http.MethodDelete)
		if err != nil {
			SendJsonResponse(w, http.StatusMethodNotAllowed, err, err.Error())
			return
		}

		uploadApp, err := b.Admin.GetApp("upload")
		if err != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}

		params := FormatRequestParameters(r, b)
		isAllowed := uploadApp.Permissions.HasPermission(params.Roles, OperationDelete)
		if !isAllowed {
			SendJsonResponse(w, http.StatusForbidden, nil, "User is not allowed to delete this resource")
			return
		}

		var instance Upload

		instanceId := GetUrlParam("id", r)

		result := b.DB.FindById(instanceId, &instance, "")
		if result.Error != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, result.Error.Error())
			return
		}

		if instance == (Upload{}) {
			SendJsonResponse(w, http.StatusNotFound, nil, "File not found")
			return
		}

		// Delete the file from disk
		err = b.Store.DeleteFile(*instance.FileData)
		if err != nil {
			log.Error().Err(err).Msg("Error deleting file")
		}

		// Delete the record from the database
		result = b.DB.Delete(&instance)
		if result.Error != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, result.Error.Error())
			return
		}

		msg := "File deleted successfully"
		if err != nil {
			msg = err.Error() + ". Database entry deleted successfuly."
		}

		SendJsonResponse(w, http.StatusOK, nil, msg)
	}
}

// getStaticHandler returns a handler function that serves static files from the
// configured folder.
func (b *Builder) GetStaticHandler(cfg *UploaderConfig) HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		requestPath := r.URL.Path
		if !strings.HasPrefix(requestPath, cfg.StaticPath) {
			log.Error().Msgf("Invalid request path: %s", requestPath)
			SendJsonResponse(w, http.StatusNotFound, nil, "File not found")
			return
		}

		fileName := strings.TrimPrefix(requestPath, cfg.StaticPath)
		filePath := filepath.Join(cfg.Folder, filepath.Base(fileName))

		bytes, err := b.Store.ReadFile(&FileData{Path: filePath})
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
	data := strings.Split(mime, "/")
	return data[0], data[1]
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
func handleUploadError(store Store, fileData FileData, w http.ResponseWriter, err error) {
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
