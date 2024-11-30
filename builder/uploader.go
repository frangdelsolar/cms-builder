package builder

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

const (
	defaultUploadFolder = "uploads"
)

type Upload struct {
	*SystemData
	FileName string `json:"fileName"`
	FilePath string `json:"filePath"`
	Url      string `json:"url"`
}

type UploaderConfig struct {
	MaxSize            int64
	Authenticate       bool
	SupportedMimeTypes []string
	Folder             string
}

// getUploadPostHandler returns a handler function that responds to POST requests
// on the uploads endpoint, e.g. /api/uploads.
//
// The handler function will save the uploaded file to disk and store the file
// information in the database. It will also handle errors and return a 400
// error if the request body is not valid JSON, or a 500 error if there is an
// error storing the file or saving the file information to the database.
func (b *Builder) GetUploadPostHandler(config *UploaderConfig) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Validate method
		err := validateRequestMethod(r, http.MethodPost)
		if err != nil {
			SendJsonResponse(w, http.StatusMethodNotAllowed, nil, err.Error())
			return
		}

		// Parse the multipart form data
		err = r.ParseMultipartForm(config.MaxSize)
		if err != nil {
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

		validContentType, err := validateContentType(contentType, config.SupportedMimeTypes)
		if err != nil {
			SendJsonResponse(w, http.StatusBadRequest, nil, err.Error())
			return
		}

		if !validContentType {
			SendJsonResponse(w, http.StatusBadRequest, nil, "Invalid file type")
			return
		}

		// Create the uploads directory if it doesn't exist
		uploadsDir := defaultUploadFolder
		if config.Folder != "" {
			uploadsDir = config.Folder
		}

		err = os.MkdirAll(uploadsDir, os.ModePerm)
		if err != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}

		// Create the file path
		fileName := randomizeFileName(header.Filename)
		path := filepath.Join(uploadsDir, fileName)

		// Save the file to disk
		dst, err := os.Create(path)
		if err != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}
		defer dst.Close()

		_, err = io.Copy(dst, file)
		if err != nil {
			handleUploadError(path, w, err)
			return
		}

		// Store filedata in db
		uploadApp, err := b.Admin.GetApp("upload")
		if err != nil {
			handleUploadError(path, w, err)
			return
		}

		uploadRequestBody := map[string]interface{}{
			"fileName": fileName,
			"filePath": path,
			"url":      "/static/" + fileName,
		}

		uploadData, err := json.Marshal(uploadRequestBody)
		if err != nil {
			handleUploadError(path, w, err)
			return
		}

		request := &http.Request{
			Method: http.MethodPost,
			Header: r.Header,
			Body:   io.NopCloser(bytes.NewBuffer(uploadData)),
		}

		// This will send the response to the client
		uploadApp.ApiNew(b.DB)(w, request)
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
func (b *Builder) GetUploadDeleteHandler(config *UploaderConfig) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		uploadApp, err := b.Admin.GetApp("upload")
		if err != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}
		userId := getRequestUserId(r, &uploadApp)

		var instance Upload
		// Query the database to find the record by ID
		result := b.DB.FindById(id, &instance, userId, true)
		if result.Error != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, result.Error.Error())
			return
		}

		if instance == (Upload{}) {
			SendJsonResponse(w, http.StatusNotFound, nil, "File not found")
			return
		}

		// Delete the file from disk
		err = os.Remove(instance.FilePath)
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
// configured folder. The handler will serve files from the configured folder
// under the path "/public/static/" if config.Authenticate is false, or
// "/private/static/" if config.Authenticate is true.
func (b *Builder) GetStaticHandler(cfg *UploaderConfig) HandlerFunc {
	prefix := config.GetString(EnvKeys.Domain) + "/" + config.GetString(EnvKeys.StaticPath) + "/"
	handler := http.StripPrefix(prefix, http.FileServer(http.Dir(cfg.Folder)))
	return func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r)
	}
}

// getMimeTypeAndExtension takes a mime string and returns the mime type and extension
// split by the "/" character. For example, "audio/wav" would return "audio" and "wav".
func getMimeTypeAndExtension(mime string) (string, string) {
	data := strings.Split(mime, "/")
	return data[0], data[1]
}

// validateContentType takes a content type and a list of supported mime types
// and returns true if the content type is supported, and false otherwise.
// It also returns an error if the content type is not supported.
//
// The supported mime types can be in different formats, such as:
// - "*": Supports all mime types.
// - "*/*": Supports all mime types with a specific mime type.
// - "audio/*": Supports all mime types with a specific mime type and extension.
// - "audio/wav": Supports a specific mime type and extension.
func validateContentType(contentType string, supportedMimeTypes []string) (bool, error) {
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

// randomizeFileName takes a file name and returns a new file name that is
// randomized using the current timestamp. The file name is also sanitized
// to replace spaces, forward slashes, and backslashes with underscores.
func randomizeFileName(fileName string) string {
	// Replace spaces with underscores
	fileName = strings.ReplaceAll(fileName, " ", "_")

	// Replace forward slashes with underscores
	fileName = strings.ReplaceAll(fileName, "/", "_")

	// Replace backslashes with underscores
	fileName = strings.ReplaceAll(fileName, "\\", "_")

	// Add the current timestamp to the file name
	now := strconv.FormatInt(time.Now().UnixNano(), 10)
	fileName = now + "_" + fileName

	return fileName
}

// DeleteFile takes a file path and deletes the file from disk.
// It returns an error if the file cannot be deleted.
func DeleteFile(path string) error {
	// Log the file path to be deleted
	log.Warn().Msgf("Deleting file: %s", path)

	// Attempt to delete the file
	if err := os.Remove(path); err != nil {
		// Log the error if the file cannot be deleted
		log.Println("Error deleting file:", err)
		return err
	}

	// Return nil if the file is successfully deleted
	return nil
}

// handleUploadError takes a file path and a writer, and an error.
// It logs the error, deletes the file from disk, and writes a JSON response
// with the error message to the writer.
func handleUploadError(filePath string, w http.ResponseWriter, err error) {
	// Log the error at the error level
	log.Error().Err(err).Msgf("Error uploading file: %s. Rolling back...", filePath)

	// Attempt to delete the file from disk
	DeleteFile(filePath)

	// Write a JSON response with the error message to the writer
	// at the internal server error (500) status code.
	SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
}
