package builder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	defaultUploadFolder = "uploads"
)

type Upload struct {
	*SystemData
	FileName string `json:"fileName"`
	FilePath string `json:"filePath"`
}

type UploaderConfig struct {
	MaxSize            int64
	Authenticate       bool
	SupportedMimeTypes []string
	Folder             string
}

func (b *Builder) getUploaderHandler(config *UploaderConfig) HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		method := r.Method
		if method == http.MethodDelete {
			handleUploadDelete(r, w, b)
		} else if method == http.MethodPost {
			handleUploadPost(r, w, config, b)
		} else {
			SendJsonResponse(w, http.StatusMethodNotAllowed, nil, fmt.Errorf("method not allowed").Error())
		}
	}
}

func handleUploadDelete(r *http.Request, w http.ResponseWriter, b *Builder) {
	id := r.FormValue("id")
	uploadApp, err := b.admin.GetApp("upload")
	if err != nil {
		SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
		return
	}
	userId := getRequestUserId(r, &uploadApp)

	var instance Upload
	// Query the database to find the record by ID
	result := b.db.FindById(id, &instance, userId, true)
	if result.Error != nil {
		SendJsonResponse(w, http.StatusInternalServerError, nil, result.Error.Error())
		return
	}

	// Delete the file from disk
	err = os.Remove(instance.FilePath)
	if err != nil {
		SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
		return
	}

	// Delete the record from the database
	result = b.db.Delete(&instance)
	if result.Error != nil {
		SendJsonResponse(w, http.StatusInternalServerError, nil, result.Error.Error())
		return
	}

	SendJsonResponse(w, http.StatusOK, nil, "File deleted successfully")

}

func handleUploadPost(r *http.Request, w http.ResponseWriter, config *UploaderConfig, b *Builder) {
	// Parse the multipart form data
	err := r.ParseMultipartForm(config.MaxSize)
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

	// Store data in db
	uploadApp, err := b.admin.GetApp("upload")
	if err != nil {
		handleUploadError(path, w, err)
		return
	}

	uploadRequestBody := map[string]interface{}{
		"fileName": header.Filename,
		"filePath": path,
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
	uploadApp.ApiNew(b.db)(w, request)
}

func getMimeTypeAndExtension(mime string) (string, string) {
	data := strings.Split(mime, "/")
	return data[0], data[1]
}

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

func randomizeFileName(fileName string) string {
	fileName = strings.ReplaceAll(fileName, " ", "_")
	fileName = strings.ReplaceAll(fileName, "/", "_")
	fileName = strings.ReplaceAll(fileName, "\\", "_")
	now := strconv.FormatInt(time.Now().UnixNano(), 10)
	fileName = now + "_" + fileName
	return fileName
}

func DeleteFile(path string) error {
	log.Info().Msgf("Deleting file: %s", path)
	if err := os.Remove(path); err != nil {
		log.Println("Error deleting file:", err)
		return err
	}

	return nil
}

func handleUploadError(filePath string, w http.ResponseWriter, err error) {
	log.Error().Err(err).Msgf("Error uploading file: %s. Rolling back...", filePath)
	DeleteFile(filePath)
	SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
}
