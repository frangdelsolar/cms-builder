package builder_test

import (
	"encoding/json"
	"net/http"
	"testing"

	builder "github.com/frangdelsolar/cms-builder/cms-builder-server"
	th "github.com/frangdelsolar/cms-builder/cms-builder-server/test_helpers"
	"github.com/stretchr/testify/assert"
)

// TestNewAdmin tests that NewAdmin returns a non-nil Admin instance.
func TestNewAdmin(t *testing.T) {
	t.Log("Testing NewAdmin")
	e, err := th.GetDefaultEngine()
	assert.NoError(t, err, "GetDefaultEngine should not return an error")

	admin := builder.NewAdmin(e.Engine)

	assert.NotNil(t, admin, "NewAdmin should return a non-nil Admin instance")
}

// TestRegisterApp tests that RegisterApp registers a new App, applies database migration,
// and registers API routes for CRUD operations.
//
// It also tests that the registered App is accessible via the GetApp method.
func TestRegisterApp(t *testing.T) {
	e, err := th.GetDefaultEngine()
	assert.NoError(t, err, "GetDefaultEngine should not return an error")

	// define a test struct to be registered
	type testStruct struct {
		*builder.SystemData
		Field string
	}

	permissions := builder.RolePermissionMap{}

	t.Log("Testing RegisterApp")
	app, err := e.Admin.Register(testStruct{}, false, permissions, nil)
	assert.NoError(t, err, "Register should not return an error")
	assert.NotNil(t, app, "Register should return a non-nil App")
	assert.Equal(t, "testStruct", app.Name(), "App name should be 'testStruct'")
	assert.Equal(t, "testStructs", app.PluralName(), "App plural name should be 'testStructs'")

	// check if the app is registered
	t.Log("Testing GetApp")
	retrievedApp, err := e.Admin.GetApp("teststruct")
	assert.Equal(t, app.Name(), retrievedApp.Name(), "GetApp should return the same app")
	assert.NoError(t, err, "GetApp should not return an error")
}

// TestRegisterAPIRoutes tests that the RegisterApp method registers the expected routes with the Server.
//
// It registers a test App with the admin instance, and then checks that the server has the expected routes.
func TestRegisterAPIRoutes(t *testing.T) {
	e, err := th.GetDefaultEngine()
	assert.NoError(t, err, "GetDefaultEngine should not return an error")

	type TwoWords struct {
		*builder.SystemData
		Field string
	}

	permissions := builder.RolePermissionMap{}

	t.Log("Testing RegisterApp")
	e.Admin.Register(TwoWords{}, false, permissions, nil)

	handler := func(w http.ResponseWriter, r *http.Request) {}

	// check ig server has expected routs
	expectedRoutes := []builder.RouteHandler{
		builder.NewRouteHandler("/api/two-words", handler, "two-words-list", true, http.MethodGet, nil),
		builder.NewRouteHandler("/api/two-words/new", handler, "two-words-new", true, http.MethodPost, TwoWords{}),
		builder.NewRouteHandler("/api/two-words/{id}", handler, "two-words-get", true, http.MethodGet, nil),
		builder.NewRouteHandler("/api/two-words/{id}/delete", handler, "two-words-delete", true, http.MethodDelete, nil),
		builder.NewRouteHandler("/api/two-words/{id}/update", handler, "two-words-update", true, http.MethodPut, TwoWords{}),
	}

	routes := e.Server.GetRoutes()
	for _, expectedRoute := range expectedRoutes {
		found := false
		for _, route := range routes {
			if route.Route == expectedRoute.Route {
				assert.Equal(t, expectedRoute.Name, route.Name, "Route name should be the same")
				assert.Equal(t, expectedRoute.RequiresAuth, route.RequiresAuth, "Route requires auth should be the same")
				assert.Equal(t, expectedRoute.Schema, route.Schema, "Route schema should be the same")
				assert.Equal(t, expectedRoute.Method, route.Method, "Route method should be the same")
				found = true
			}
		}

		assert.True(t, found)
	}
}

// TestUserCanRetrieveAllowedResources tests that a user can retrieve a resource if they have the correct permissions.
//
// It creates a resource for the logged-in user, and then checks that the response contains the same record.
func TestUserCanRetrieveAllowedResources(t *testing.T) {
	e, err := th.GetDefaultEngine()
	assert.NoError(t, err, "GetDefaultEngine should not return an error")

	// Create a resource for some user
	instance, user, userRollback := th.CreateMockResource(t, e.DB, e.App, nil)
	defer userRollback()

	// Create a helper request to get the detail
	request, _, _ := th.NewRequest(
		http.MethodGet,
		"",
		true,
		user,
		map[string]string{"id": instance.GetIDString()},
	)

	t.Log("Getting the detail for user")
	var result th.MockStruct
	response, err := th.ExecuteApiCall(
		t,
		e.App.ApiDetail(e.DB),
		request,
		&result,
	)

	assert.NoError(t, err, "ApiDetail should not return an error")
	assert.NotNil(t, response, "ApiDetail should return a non-nil response")

	assert.Equal(t, instance.ID, result.ID, "ID should be the same")
	assert.Equal(t, instance.Field, result.Field, "Field should be the same")
	assert.Equal(t, instance.CreatedByID, result.CreatedByID, "CreatedBy should be the same")
}

// TestUserCanNotRetrieveDeniedResources tests that a user can not retrieve a resource if they don't have the correct permissions.
//
// It creates two resources for two different users, and then checks that each user can not retrieve the resource of the other user.
func TestUserCanNotRetrieveDeniedResources(t *testing.T) {
	e, err := th.GetDefaultEngine()
	assert.NoError(t, err, "GetDefaultEngine should not return an error")

	// Create a resource for user A
	instanceA, userA, userARollback := th.CreateMockResource(t, e.DB, e.App, nil)
	defer userARollback()

	// Create a resource for user B
	instanceB, userB, userBRollback := th.CreateMockResource(t, e.DB, e.App, nil)
	defer userBRollback()

	// User A tries to get the detail for user A
	t.Log("User A tries to get the detail for user B")
	requestA, _, _ := th.NewRequest(
		http.MethodGet,
		"",
		true,
		userA,
		map[string]string{"id": instanceB.GetIDString()},
	)

	responseA, err := th.ExecuteApiCall(t, e.App.ApiDetail(e.DB), requestA, nil)
	assert.Error(t, err, "ApiDetail should return an error")
	assert.False(t, responseA.Success, "ParseResponse should return an error response")
	assert.Contains(t, responseA.Message, "record not found", "The response should be an error")

	// User B tries to get the detail for user A
	requestB, _, _ := th.NewRequest(
		http.MethodGet,
		"",
		true,
		userB,
		map[string]string{"id": instanceA.GetIDString()},
	)

	responseB, err := th.ExecuteApiCall(t, e.App.ApiDetail(e.DB), requestB, nil)
	assert.Error(t, err, "ApiDetail should return an error")
	assert.False(t, responseB.Success, "ParseResponse should return an error response")
	assert.Contains(t, responseB.Message, "record not found", "The response should be an error")
}

// TestUserCanListAllowedResources tests that a user can list resources if they have the correct permissions.
//
// It creates two resources for some user, and then checks that the response contains the two resources.
func TestUserCanListAllowedResources(t *testing.T) {
	e, err := th.GetDefaultEngine()
	assert.NoError(t, err, "GetDefaultEngine should not return an error")

	// Create two resource for some user
	instanceA, user, userRollback := th.CreateMockResource(t, e.DB, e.App, nil)
	instanceB, _, userbRollback := th.CreateMockResource(t, e.DB, e.App, user)
	defer userRollback()
	defer userbRollback()

	// Create a helper request to get the list
	request, _, _ := th.NewRequest(
		http.MethodGet,
		"",
		true,
		user,
		nil,
	)

	t.Log("Getting the List for user", request)
	var result []th.MockStruct
	response, err := th.ExecuteApiCall(
		t,
		e.App.ApiList(e.DB),
		request,
		&result,
	)

	assert.NoError(t, err, "ApiList should not return an error")
	assert.NotNil(t, response, "ApiList should return a non-nil response")
	assert.Equal(t, 2, len(result), "List should contain two items")

	// results are order by id desc, thats why we reverse
	resultA := result[1]
	resultB := result[0]

	assert.Equal(t, instanceA.ID, resultA.ID, "ID should be the same")
	assert.Equal(t, instanceA.Field, resultA.Field, "Field should be the same")
	assert.Equal(t, instanceA.CreatedByID, resultA.CreatedByID, "CreatedBy should be the same")

	assert.Equal(t, instanceB.ID, resultB.ID, "ID should be the same")
	assert.Equal(t, instanceB.Field, resultB.Field, "Field should be the same")
	assert.Equal(t, instanceB.CreatedByID, resultB.CreatedByID, "CreatedBy should be the same")
}

// TestUserCanNotListDeniedResources tests that a user can not list resources if they don't have the correct permissions.
//
// It creates two resources for some user, and then checks that the response contains an empty list.
func TestUserCanNotListDeniedResources(t *testing.T) {
	e, err := th.GetDefaultEngine()
	assert.NoError(t, err, "GetDefaultEngine should not return an error")

	// Create two resource for some user
	t.Log("Creating two resources for userA")
	_, userA, userARollback := th.CreateMockResource(t, e.DB, e.App, nil)
	th.CreateMockResource(t, e.DB, e.App, userA)
	defer userARollback()

	// Create a helper request to get the list for user B
	request, _, userBRollback := th.NewRequest(
		http.MethodGet,
		"",
		true,
		nil,
		nil,
	)
	defer userBRollback()

	t.Log("Getting the List for userB")
	var result []th.MockStruct
	response, err := th.ExecuteApiCall(
		t,
		e.App.ApiList(e.DB),
		request,
		&result,
	)

	assert.NoError(t, err, "ApiList should not return an error")
	assert.NotNil(t, response, "ApiList should return a non-nil response")
	assert.Equal(t, 0, len(result), "List should be empty")
}

// TestUserCanCreateAllowedResources tests that a user can create a resource if they have the correct permissions.
//
// It creates a new resource and checks that the response contains the created resource and that the resource is persisted in the database.
func TestUserCanCreateAllowedResources(t *testing.T) {
	e, err := th.GetDefaultEngine()
	assert.NoError(t, err, "GetDefaultEngine should not return an error")

	// Create a resource for some user
	instance, user, userRollback := th.CreateMockResource(t, e.DB, e.App, nil)
	defer userRollback()

	assert.NotNil(t, instance.ID, "ID should not be nil")
	assert.NotNil(t, instance.Field, "Field should not be nil")
	assert.NotNil(t, instance.CreatedByID, "CreatedBy should not be nil")

	// Validate that historyEntry is created
	historyEntry, err := builder.GetHistoryEntryForInstanceFromDB(e.DB, user.GetIDString(), instance, instance.GetIDString(), "mockstruct", builder.CreateCRUDAction)
	assert.NoError(t, err, "GetHistoryEntryForInstanceFromDB should not return an error")
	assert.NotNil(t, historyEntry, "HistoryEntry should not be nil")
}

// TestUserCanNotCreateDeniedResources tests that a user cannot create a resource if they do not have the correct permissions.
//
// It creates a new resource and checks that the response contains an error message indicating that the user does not have the correct permissions.
func TestUserCanNotCreateDeniedResources(t *testing.T) {
	e, err := th.GetDefaultEngine()
	assert.NoError(t, err, "GetDefaultEngine should not return an error")

	// Create a resource without authentication
	request, user, userRollback := th.NewRequest(
		http.MethodPost,
		`{"field": "test"}`,
		false,
		nil,
		nil,
	)
	defer userRollback()

	assert.Nil(t, user, "User should be nil")

	t.Log("Creating a resource without authentication")
	var result th.MockStruct
	response, err := th.ExecuteApiCall(
		t,
		e.App.ApiCreate(e.DB),
		request,
		&result,
	)
	assert.Equal(t, result, th.MockStruct{}, "Result should be empty")
	assert.NoError(t, err, "ApiNew should not return an error")
	assert.NotNil(t, response, "ApiNew should return a non-nil response")
	assert.Contains(t, response.Message, "not allowed", "The response should be an error message")
}

// TestUserCanUpdateAllowedResources tests that a user can update a resource if they have the correct permissions.
//
// It creates a resource for the logged-in user, and then checks that the response contains the updated record.
func TestUserCanUpdateAllowedResources(t *testing.T) {
	e, err := th.GetDefaultEngine()
	assert.NoError(t, err, "GetDefaultEngine should not return an error")

	// Create a resource for the logged-in user
	instance, user, userRollback := th.CreateMockResource(t, e.DB, e.App, nil)
	defer userRollback()

	// Update the resource
	request, _, _ := th.NewRequest(
		http.MethodPut,
		`{"field": "updated_field"}`,
		true,
		user,
		map[string]string{"id": instance.GetIDString()},
	)

	var updatedInstance th.MockStruct
	response, err := th.ExecuteApiCall(t, e.App.ApiUpdate(e.DB), request, &updatedInstance)

	assert.NoError(t, err, "ApiUpdate should not return an error")
	assert.NotNil(t, response, "ApiUpdate should return a non-nil response")

	// Verify the update was successful
	assert.Equal(t, instance.ID, updatedInstance.ID, "ID should be the same")
	assert.Equal(t, "updated_field", updatedInstance.Field, "Field should be the same")
	assert.Equal(t, user.ID, updatedInstance.CreatedByID, "CreatedBy should be the same")
	assert.Equal(t, user.ID, updatedInstance.UpdatedByID, "UpdatedBy should be the same")

	// Validate that historyEntry is created
	historyEntry, err := builder.GetHistoryEntryForInstanceFromDB(e.DB, user.GetIDString(), updatedInstance, updatedInstance.GetIDString(), "mockstruct", builder.UpdateCRUDAction)
	assert.NoError(t, err, "GetHistoryEntryForInstanceFromDB should not return an error")
	assert.NotNil(t, historyEntry, "HistoryEntry should not be nil")
}

// TestUserCanNotUpdateDeniedResources tests that a user can not update a resource if they don't have the correct permissions.
//
// It creates a resource for the logged-in user, and then checks that the response contains an error.
func TestUserCanNotUpdateDeniedResources(t *testing.T) {
	e, err := th.GetDefaultEngine()
	assert.NoError(t, err, "GetDefaultEngine should not return an error")

	// Create a resource for the logged-in user
	instance, _, userARollback := th.CreateMockResource(t, e.DB, e.App, nil)
	defer userARollback()

	// Update the resource with a different user
	request, _, userBRollback := th.NewRequest(
		http.MethodPut,
		`{"field": "updated_field"}`,
		true,
		nil,
		map[string]string{"id": instance.GetIDString()},
	)
	defer userBRollback()

	response, err := th.ExecuteApiCall(t, e.App.ApiUpdate(e.DB), request, nil)

	assert.Error(t, err, "ApiUpdate should return an error")
	assert.NotNil(t, response, "ApiUpdate should return a non-nil response")
	assert.Contains(t, response.Message, "record not found", "The response should be an error")
}

// TestUserCanDeleteAllowedResources tests that a user can delete a resource if they have the correct permissions.
//
// It creates a resource for the logged-in user, and then checks that the response contains an error message indicating that the user does not have the correct permissions.
func TestUserCanDeleteAllowedResources(t *testing.T) {
	e, err := th.GetDefaultEngine()
	assert.NoError(t, err, "GetDefaultEngine should not return an error")

	// Create a resource for the logged-in user
	instance, user, userRollback := th.CreateMockResource(t, e.DB, e.App, nil)
	defer userRollback()

	// Update the resource
	deleteRequest, _, _ := th.NewRequest(
		http.MethodDelete,
		"",
		true,
		user,
		map[string]string{"id": instance.GetIDString()},
	)

	t.Log("Deleting the resource")
	var resultA th.MockStruct
	responseA, err := th.ExecuteApiCall(t, e.App.ApiDelete(e.DB), deleteRequest, &resultA)

	assert.NoError(t, err, "ApiDelete should not return an error")
	assert.NotNil(t, responseA, "ApiDelete should return a non-nil response")
	assert.Equal(t, responseA.Success, true, "Success should be true")

	// Create a helper request to get the detail
	getRequest, _, _ := th.NewRequest(
		http.MethodGet,
		"",
		true,
		user,
		map[string]string{"id": instance.GetIDString()},
	)

	t.Log("Getting the detail for user")
	var resultB th.MockStruct
	responseB, err := th.ExecuteApiCall(
		t,
		e.App.ApiDetail(e.DB),
		getRequest,
		&resultB,
	)

	assert.NoError(t, err, "ApiDetail should not return an error")
	assert.NotNil(t, responseB, "ApiDetail should return a non-nil response")
	assert.Equal(t, responseB.Success, false, "Success should be false")
	assert.Contains(t, responseB.Message, "record not found", "The response should be an error")
	assert.Equal(t, resultB, th.MockStruct{}, "The result should be empty")

	// Validate that historyEntry is created
	historyEntry, err := builder.GetHistoryEntryForInstanceFromDB(e.DB, user.GetIDString(), instance, instance.GetIDString(), "mockstruct", builder.DeleteCRUDAction)
	assert.NoError(t, err, "GetHistoryEntryForInstanceFromDB should not return an error")
	assert.NotNil(t, historyEntry, "HistoryEntry should not be nil")
}

// TestUserCanNotDeleteDeniedResources tests that a user cannot delete a resource if they don't have the correct permissions.
//
// It creates a resource for the logged-in user, and then checks that the response contains an error message indicating that the user does not have the correct permissions.
func TestUserCanNotDeleteDeniedResources(t *testing.T) {
	e, err := th.GetDefaultEngine()
	assert.NoError(t, err, "GetDefaultEngine should not return an error")

	// Create a resource for the logged-in user
	instance, userA, userARollback := th.CreateMockResource(t, e.DB, e.App, nil)
	defer userARollback()

	// Update the resource
	deleteRequest, _, userBRollback := th.NewRequest(
		http.MethodDelete,
		"",
		true,
		nil,
		map[string]string{"id": instance.GetIDString()},
	)
	defer userBRollback()

	t.Log("Attempting to delete the resource with userB")
	var resultA th.MockStruct
	responseA, err := th.ExecuteApiCall(t, e.App.ApiDelete(e.DB), deleteRequest, &resultA)

	assert.NoError(t, err, "ApiDetail should not return an error")
	assert.NotNil(t, responseA, "ApiDetail should return a non-nil response")
	assert.Equal(t, responseA.Success, false, "Success should be false")
	assert.Contains(t, responseA.Message, "record not found", "The response should be an error")
	assert.Equal(t, resultA, th.MockStruct{}, "The result should be empty")

	// Create a helper request to get the detail
	t.Log("Validating userA can retrieve the resource")
	getRequest, _, _ := th.NewRequest(
		http.MethodGet,
		"",
		true,
		userA,
		map[string]string{"id": instance.GetIDString()},
	)

	t.Log("Getting the detail for userA")
	var resultB th.MockStruct
	responseB, err := th.ExecuteApiCall(
		t,
		e.App.ApiDetail(e.DB),
		getRequest,
		&resultB,
	)

	assert.NoError(t, err, "ApiDetail should not return an error")
	assert.NotNil(t, responseB, "ApiDetail should return a non-nil response")
	assert.Equal(t, responseB.Success, true, "Success should be true")

	assert.Equal(t, instance.ID, resultB.ID, "ID should be the same")
	assert.Equal(t, instance.Field, resultB.Field, "Field should be the same")
	assert.Equal(t, instance.CreatedByID, resultB.CreatedByID, "CreatedBy should be the same")
}

// TestValidators tests that a user cannot create a resource with invalid
// values. It creates a new resource and checks that the response contains an
// error message indicating that the validation failed.
func TestCreateCallsValidators(t *testing.T) {
	e, err := th.GetDefaultEngine()
	assert.NoError(t, err, "GetDefaultEngine should not return an error")

	// Create a resource
	request, _, rollback := th.NewRequest(
		http.MethodPost,
		`{"field": ""}`,
		true,
		nil,
		nil,
	)
	defer rollback()

	t.Log("Creating a resource")
	var result th.MockStruct
	response, err := th.ExecuteApiCall(
		t,
		e.App.ApiCreate(e.DB),
		request,
		&result,
	)

	assert.NoError(t, err, "ExecuteApiCall should not return an error")
	assert.False(t, response.Success, "ApiDetail should return an error")
	assert.Contains(t, response.Message, "Validation failed", "The response should be an error")
}

// TestUpdateCallsValidators tests that a user cannot update a resource with
// invalid values. It creates a resource, tries to update it with invalid values,
// and checks that the response contains an error message indicating that the
// validation failed.
func TestUpdateCallsValidators(t *testing.T) {
	e, err := th.GetDefaultEngine()
	assert.NoError(t, err, "GetDefaultEngine should not return an error")

	// Create a resource
	instance, user, rollback := th.CreateMockResource(t, e.DB, e.App, nil)
	defer rollback()

	// Update the resource
	request, _, _ := th.NewRequest(
		http.MethodPut,
		`{"field": ""}`,
		true,
		user,
		map[string]string{"id": instance.GetIDString()},
	)
	defer rollback()

	t.Log("Trying to update the resource")
	var result th.MockStruct
	response, _ := th.ExecuteApiCall(
		t,
		e.App.ApiUpdate(e.DB),
		request,
		&result,
	)

	// assert.NoError(t, err, "ExecuteApiCall should not return an error")
	assert.False(t, response.Success, "ApiDetail should return an error")
	assert.Contains(t, response.Message, "Validation failed", "The response should be an error")
}

// TestUserCanNotReplaceCreatedByIDOnCreate tests that a user cannot create a resource with a createdById or updatedById that is not their own user ID.
func TestUserCanNotReplaceCreatedByIDOnCreate(t *testing.T) {
	e, err := th.GetDefaultEngine()
	assert.NoError(t, err, "GetDefaultEngine should not return an error")

	// Create a resource for the logged-in user
	request, user, rollback := th.NewRequest(
		http.MethodPost,
		`{"field": "some value", "createdById": 9999991, "updatedById": 9999991}`,
		true,
		nil,
		nil,
	)
	defer rollback()

	t.Log("Creating a resource")

	var instance th.MockStruct
	response, err := th.ExecuteApiCall(t, e.App.ApiCreate(e.DB), request, &instance)
	assert.NoError(t, err, "ExecuteApiCall should not return an error")
	assert.NotNil(t, response, "ApiNew should return a non-nil response")

	assert.Equal(t, user.ID, instance.CreatedByID, "CreatedByID should be the logged in user")
	assert.Equal(t, user.ID, instance.UpdatedByID, "UpdatedByID should be the logged in user")
}

// TestUserCanNotReplaceCreatedByIDOnUpdate tests that a user cannot update a resource with a createdById or updatedById that is not their own user ID.
func TestUserCanNotReplaceCreatedByIDOnUpdate(t *testing.T) {
	e, err := th.GetDefaultEngine()
	assert.NoError(t, err, "GetDefaultEngine should not return an error")

	// Create a resource for the logged-in user
	instance, user, userRollback := th.CreateMockResource(t, e.DB, e.App, nil)
	defer userRollback()

	// Update the resource
	request, _, _ := th.NewRequest(
		http.MethodPut,
		`{"createdById": 9999999, "updatedById": 9999999}`,
		true,
		user,
		map[string]string{"id": instance.GetIDString()},
	)

	t.Log("Updating the resource systemData")
	var updatedInstance th.MockStruct
	response, _ := th.ExecuteApiCall(t, e.App.ApiUpdate(e.DB), request, &updatedInstance)

	assert.NotNil(t, response, "ApiUpdate should return a non-nil response")
	assert.Equal(t, true, response.Success, "ApiUpdate should return a success response")

	// Verify the update was successful
	assert.Equal(t, instance.ID, updatedInstance.ID, "ID should be the same")
	assert.Equal(t, user.ID, updatedInstance.CreatedByID, "CreatedByID should be the logged in user")
	assert.Equal(t, user.ID, updatedInstance.UpdatedByID, "UpdatedByID should be the logged in user")
}

// TestUserCanNotReplaceInstanceIDOnUpdate tests that a user cannot update a resource with an instance ID that is not the same as the one in the request.
func TestUserCanNotReplaceInstanceIDOnUpdate(t *testing.T) {
	e, err := th.GetDefaultEngine()
	assert.NoError(t, err, "GetDefaultEngine should not return an error")

	// Create a resource for the logged-in user
	instance, user, userRollback := th.CreateMockResource(t, e.DB, e.App, nil)
	defer userRollback()

	// Update the resource
	request, _, _ := th.NewRequest(
		http.MethodPut,
		`{"id": 79999999}`,
		true,
		user,
		map[string]string{"id": instance.GetIDString()},
	)
	t.Log("Updating the resource with an some ID")
	var updatedInstance th.MockStruct
	response, _ := th.ExecuteApiCall(t, e.App.ApiUpdate(e.DB), request, &updatedInstance)
	assert.NotNil(t, response, "ApiUpdate should return a non-nil response")
	assert.Equal(t, true, response.Success, "ApiUpdate should return a success response")
	// Verify the update was successful
	assert.Equal(t, instance.GetIDString(), updatedInstance.GetIDString(), "ID should remain the same")
	assert.Equal(t, user.ID, updatedInstance.CreatedByID, "CreatedByID should be the logged in user")
	assert.Equal(t, user.ID, updatedInstance.UpdatedByID, "UpdatedByID should be the logged in user")
}

// TestJsonifyInterface tests that JsonifyInterface correctly converts a struct to a map[string]interface{}
func TestJsonifyInterface(t *testing.T) {

	t.Log("Testing JsonifyInterface with omitempty. Not Working")
	t.Skip()
	//TODO: find a way for this to work
	type JsonTest struct {
		Field      string `json:"field"`
		NotOmitted string `json:"notOmitted"`
	}

	testStruct := JsonTest{}
	// Convert the struct to a map
	data, err := builder.JsonifyInterface(&testStruct)
	assert.NoError(t, err, "JsonifyInterface should not return an error")

	// Marshal the map to JSON
	bytes, err := json.Marshal(data)
	assert.NoError(t, err, "json.Marshal should not return an error")

	// Assert the expected output
	assert.Equal(t, `{"field":"","notOmitted":""}`, string(bytes))

}

// TestJsonifyInterfaceWithOmitempty tests that JsonifyInterface correctly converts a struct to a map[string]interface{}, including the use of "omitempty" tags.
func TestJsonifyInterfaceWithOmitempty(t *testing.T) {

	// t.Log("Testing JsonifyInterface with omitempty. Not Working")
	t.Skip()
	//TODO: find a way for this to work
	type JsonTest struct {
		Field   string `json:"field"`
		Omitted string `json:"omitted,omitempty"`
	}

	testStruct := JsonTest{}
	// Convert the struct to a map
	data, err := builder.JsonifyInterface(&testStruct)
	assert.NoError(t, err, "JsonifyInterface should not return an error")

	// Marshal the map to JSON
	bytes, err := json.Marshal(data)
	assert.NoError(t, err, "json.Marshal should not return an error")

	// Assert the expected output
	assert.Equal(t, `{"field":"","omitted":""}`, string(bytes))

}
