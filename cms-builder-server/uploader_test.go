package builder_test

import (
	"net/http"
	"testing"

	builder "github.com/frangdelsolar/cms-builder/cms-builder-server"
	th "github.com/frangdelsolar/cms-builder/cms-builder-server/test_helpers"
	"github.com/stretchr/testify/assert"
)

const testFilePath = "test_data/test.json"

func TestUploaderGetsCreated(t *testing.T) {
	e, err := th.GetDefaultEngine()
	assert.NoError(t, err, "GetDefaultEngine should not return an error")

	app, err := e.Admin.GetApp("file")
	assert.NoError(t, err, "GetApp should not return an error")
	assert.NotNil(t, app, "GetApp should return a non-nil App")

	handler := func(w http.ResponseWriter, r *http.Request) {}

	expectedRoutes := []builder.RouteHandler{
		builder.NewRouteHandler("/api/files", handler, "files-list", true, http.MethodGet, nil),
		builder.NewRouteHandler("/api/files/{id}", handler, "files-get", true, http.MethodGet, nil),
		builder.NewRouteHandler("/api/files/new", handler, "files-new", true, http.MethodPost, nil),
		builder.NewRouteHandler("/api/files/{id}/delete", handler, "files-delete", true, http.MethodDelete, nil),
		builder.NewRouteHandler("/api/files/{id}/download", handler, "files-download", true, http.MethodGet, nil),
		builder.NewRouteHandler("/api/files/{id}/update", handler, "files-update", true, http.MethodGet, nil),
	}

	routes := e.Server.GetRoutes()
	for _, expectedRoute := range expectedRoutes {
		found := false
		for _, route := range routes {
			if route.Route == expectedRoute.Route {
				assert.Equal(t, expectedRoute.Name, route.Name, "Route name should be the same")
				assert.Equal(t, expectedRoute.RequiresAuth, route.RequiresAuth, "Route requires auth should be the same")
				found = true
			}
		}

		assert.True(t, found, "Expected route not found: %s", expectedRoute.Route)
	}

	assert.NotNil(t, app.Api, "Api handlers should not be nil")
	assert.NotNil(t, app.Api.Create, "Create handlers should not be nil")
	assert.NotNil(t, app.Api.Delete, "Delete handlers should not be nil")
	assert.NotNil(t, app.Api.Detail, "Detail handlers should not be nil")
	assert.NotNil(t, app.Api.List, "List handlers should not be nil")
	assert.NotNil(t, app.Api.Update, "Update handlers should not be nil")
}

func TestAuthenticatedCanUploadAndDeleteAllowed(t *testing.T) {
	e, err := th.GetDefaultEngine()
	assert.NoError(t, err, "GetDefaultEngine should not return an error")

	// Create a helper request to get the detail
	request, user, rollback := th.NewRequestWithFile(
		http.MethodPost,
		"",
		testFilePath,
		true,
		nil,
		nil,
	)
	defer rollback()

	app, err := e.Admin.GetApp("file")
	assert.NoError(t, err, "GetApp should not return an error")
	assert.NotNil(t, app, "GetApp should return a non-nil App")

	var result builder.File
	response, err := th.ExecuteApiCall(
		t,
		app.Api.Create(&app, e.Engine.DB),
		request,
		&result,
	)

	assert.NoError(t, err, "ApiCreate should not return an error")
	assert.NotNil(t, response, "ApiCreate should return a non-nil response")
	assert.Equal(t, response.Success, true, "Success should be true")

	assert.NotNil(t, result.ID, "ID should be something", result.ID)
	assert.NotNil(t, result.Name, "FileName should be something", result.Name)
	assert.NotNil(t, result.Path, "FilePath should be something", result.Path)
	assert.NotNil(t, result.Url, "Url should be something", result.Url)

	// clean up
	request, _, _ = th.NewRequest(
		http.MethodDelete,
		"",
		true,
		user,
		map[string]string{"id": result.GetIDString()},
	)

	response, err = th.ExecuteApiCall(
		t,
		app.ApiDelete(e.Engine.DB),
		request,
		&result,
	)

	assert.NoError(t, err, "ApiDelete should not return an error")
	assert.NotNil(t, response, "ApiDelete should return a non-nil response")
	assert.Equal(t, response.Success, true, "Success should be true")
}

func TestAnonymousCanNotUploadForbidden(t *testing.T) {
	e, err := th.GetDefaultEngine()
	assert.NoError(t, err, "GetDefaultEngine should not return an error")

	// Create a helper request to get the detail
	request, _, rollback := th.NewRequestWithFile(
		http.MethodPost,
		"",
		testFilePath,
		false,
		nil,
		nil,
	)
	defer rollback()

	var result builder.File

	app, err := e.Admin.GetApp("file")
	assert.NoError(t, err, "GetApp should not return an error")
	assert.NotNil(t, app, "GetApp should return a non-nil App")

	response, err := th.ExecuteApiCall(
		t,
		app.Api.Create(&app, e.Engine.DB),
		request,
		&result,
	)

	assert.NoError(t, err, "Upload should not return an error")
	assert.Equal(t, response.Success, false, "Success should be false")
	assert.Contains(t, response.Message, "not allowed", "Error should contain not allowed")

	assert.Equal(t, result, (builder.File{}), "Result should be nil", result)
}
