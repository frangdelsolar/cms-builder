package builder_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/frangdelsolar/cms/builder"
	th "github.com/frangdelsolar/cms/builder/test_helpers"
	"github.com/stretchr/testify/assert"
)

type MockStruct struct {
	*builder.SystemData
	Field string
}

func FieldValidator(inName interface{}) builder.FieldValidationError {
	fieldValue := fmt.Sprint(inName)
	output := builder.NewFieldValidationError("field")
	if fieldValue == "" {
		output.Error = "field cannot be empty"
		return output
	}

	return builder.FieldValidationError{}
}

func setupTest(t *testing.T) (engine *builder.Builder, admin *builder.Admin, db *builder.Database, server *builder.Server, app *builder.App, deregisterApp func()) {

	engine = th.GetDefaultEngine()
	admin, err := engine.GetAdmin()
	assert.NoError(t, err)
	assert.NotNil(t, admin)

	db, err = engine.GetDatabase()
	assert.NoError(t, err)
	assert.NotNil(t, db)

	server, err = engine.GetServer()
	assert.NoError(t, err)
	assert.NotNil(t, server)

	mockApp, err := admin.Register(MockStruct{}, false)
	assert.NoError(t, err, "Register should not return an error")

	mockApp.RegisterValidator("field", FieldValidator)

	callback := func() {
		admin.Unregister(app.Name())
	}

	return engine, admin, db, server, &mockApp, callback
}

// createTestResource creates a new resource for the given user and returns the created resource, the user, and a function to roll back the resource creation.
//
// It creates a new request with a random string as the value for the "field" key, then calls the app's ApiNew method to create a new resource. It then unmarshalls the response into the createdItem variable and returns it along with the user and the rollback function.
func createTestResource(t *testing.T, db *builder.Database, app *builder.App, user *builder.User) (*MockStruct, *builder.User, func()) {
	responseWriter := th.MockWriter{}

	data := "{\"field\": \"" + th.RandomString(10) + "\"}"
	request, user, rollback := th.NewRequest(http.MethodPost, data, true, user, nil)

	t.Logf("Creating new resource for user: %v", user.ID)
	app.ApiNew(db)(&responseWriter, request)

	var createdItem MockStruct
	err := json.Unmarshal([]byte(responseWriter.GetWrittenData()), &createdItem)
	assert.NoError(t, err, "Unmarshal should not return an error")

	return &createdItem, user, rollback
}

// executeApiCall executes an API handler function and returns the response as a string.
//
// The given request is used as the input to the handler function.
//
// The response is returned as a string, which is the JSON representation of the response.
//
// This function is meant to be used in tests.
func executeApiCall(t *testing.T, apiCall builder.HandlerFunc, request *http.Request) string {
	t.Log("Executing API call", request.Method, request.Body, request.Header)
	writer := th.MockWriter{}
	apiCall(&writer, request)
	return writer.GetWrittenData()
}

// TestNewAdmin tests that NewAdmin returns a non-nil Admin instance.
func TestNewAdmin(t *testing.T) {
	t.Log("Testing NewAdmin")
	engine := th.GetDefaultEngine()
	db, _ := engine.GetDatabase()
	server, _ := engine.GetServer()

	admin := builder.NewAdmin(db, server)

	assert.NotNil(t, admin, "NewAdmin should return a non-nil Admin instance")
}

// TestRegisterApp tests that RegisterApp registers a new App, applies database migration,
// and registers API routes for CRUD operations.
//
// It also tests that the registered App is accessible via the GetApp method.
func TestRegisterApp(t *testing.T) {
	_, admin, _, _, _, _ := setupTest(t)

	// define a test struct to be registered
	type testStruct struct {
		*builder.SystemData
		Field string
	}

	t.Log("Testing RegisterApp")
	app, err := admin.Register(testStruct{}, false)
	assert.NoError(t, err, "Register should not return an error")
	assert.NotNil(t, app, "Register should return a non-nil App")
	assert.Equal(t, "teststruct", app.Name(), "App name should be 'teststruct'")
	assert.Equal(t, "teststructs", app.PluralName(), "App plural name should be 'teststructs'")

	// check if the app is registered
	t.Log("Testing GetApp")
	retrievedApp, err := admin.GetApp("teststruct")
	assert.Equal(t, app.Name(), retrievedApp.Name(), "GetApp should return the same app")
	assert.NoError(t, err, "GetApp should not return an error")
}

// TestRegisterAPIRoutes tests that the RegisterApp method registers the expected routes with the Server.
//
// It registers a test App with the admin instance, and then checks that the server has the expected routes.
func TestRegisterAPIRoutes(t *testing.T) {
	_, admin, _, server, _, _ := setupTest(t)

	type Test struct {
		*builder.SystemData
		Field string
	}

	t.Log("Testing RegisterApp")
	admin.Register(Test{}, false)

	handler := func(w http.ResponseWriter, r *http.Request) {}

	// check ig server has expected routs
	expectedRoutes := []builder.RouteHandler{
		builder.NewRouteHandler("/api/tests", handler, "test-list", true),
		builder.NewRouteHandler("/api/tests/new", handler, "test-new", true),
		builder.NewRouteHandler("/api/tests/{id}", handler, "test-get", true),
		builder.NewRouteHandler("/api/tests/{id}/delete", handler, "test-delete", true),
		builder.NewRouteHandler("/api/tests/{id}/update", handler, "test-update", true),
	}

	routes := server.GetRoutes()
	for _, expectedRoute := range expectedRoutes {
		found := false
		for _, route := range routes {
			if route.Route == expectedRoute.Route {
				assert.Equal(t, expectedRoute.Name, route.Name, "Route name should be the same")
				assert.Equal(t, expectedRoute.RequiresAuth, route.RequiresAuth, "Route requires auth should be the same")
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
	_, _, db, _, app, deregisterApp := setupTest(t)
	defer deregisterApp()

	// Create a resource for some user
	instance, user, userRollback := createTestResource(t, db, app, nil)
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
	response := executeApiCall(
		t,
		app.ApiDetail(db),
		request,
	)

	var result MockStruct
	json.Unmarshal([]byte(response), &result)

	assert.Equal(t, instance.ID, result.ID, "ID should be the same")
	assert.Equal(t, instance.Field, result.Field, "Field should be the same")
	assert.Equal(t, instance.CreatedByID, result.CreatedByID, "CreatedBy should be the same")
}

// TestUserCanNotRetrieveDeniedResources tests that a user can not retrieve a resource if they don't have the correct permissions.
//
// It creates two resources for two different users, and then checks that each user can not retrieve the resource of the other user.
func TestUserCanNotRetrieveDeniedResources(t *testing.T) {
	_, _, db, _, app, deregisterApp := setupTest(t)
	defer deregisterApp()

	// Create a resource for user A
	instanceA, userA, userARollback := createTestResource(t, db, app, nil)
	defer userARollback()

	// Create a resource for user B
	instanceB, userB, userBRollback := createTestResource(t, db, app, nil)
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

	responseA := executeApiCall(t, app.ApiDetail(db), requestA)
	assert.Contains(t, responseA, "record not found", "The response should be an error")

	// User B tries to get the detail for user A
	requestB, _, _ := th.NewRequest(
		http.MethodGet,
		"",
		true,
		userB,
		map[string]string{"id": instanceA.GetIDString()},
	)

	responseB := executeApiCall(t, app.ApiDetail(db), requestB)
	assert.Contains(t, responseB, "record not found", "The response should be an error")
}

// TestUserCanListAllowedResources tests that a user can list resources if they have the correct permissions.
//
// It creates two resources for some user, and then checks that the response contains the two resources.
func TestUserCanListAllowedResources(t *testing.T) {
	_, _, db, _, app, deregisterApp := setupTest(t)
	defer deregisterApp()

	// Create two resource for some user
	instanceA, user, userRollback := createTestResource(t, db, app, nil)
	instanceB, _, _ := createTestResource(t, db, app, user)
	defer userRollback()

	// Create a helper request to get the list
	request, _, _ := th.NewRequest(
		http.MethodGet,
		"",
		true,
		user,
		nil,
	)

	t.Log("Getting the List for user")
	response := executeApiCall(
		t,
		app.ApiList(db),
		request,
	)

	var result []MockStruct
	json.Unmarshal([]byte(response), &result)

	assert.Equal(t, 2, len(result), "List should contain two items")

	resultA := result[0]
	resultB := result[1]

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
	_, _, db, _, app, deregisterApp := setupTest(t)
	defer deregisterApp()

	// Create two resource for some user
	t.Log("Creating two resources for userA")
	_, userA, userARollback := createTestResource(t, db, app, nil)
	createTestResource(t, db, app, userA)
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
	response := executeApiCall(
		t,
		app.ApiList(db),
		request,
	)

	var result []MockStruct
	json.Unmarshal([]byte(response), &result)
	assert.Equal(t, 0, len(result), "List should be empty")
}

// TestUserCanCreateAllowedResources tests that a user can create a resource if they have the correct permissions.
//
// It creates a new resource and checks that the response contains the created resource and that the resource is persisted in the database.
func TestUserCanCreateAllowedResources(t *testing.T) {
	_, _, db, _, app, deregisterApp := setupTest(t)
	defer deregisterApp()

	// Create a resource for some user
	instance, _, userRollback := createTestResource(t, db, app, nil)
	defer userRollback()

	assert.NotNil(t, instance.ID, "ID should not be nil")
	assert.NotNil(t, instance.Field, "Field should not be nil")
	assert.NotNil(t, instance.CreatedByID, "CreatedBy should not be nil")
}

// TestUserCanNotCreateDeniedResources tests that a user cannot create a resource if they do not have the correct permissions.
//
// It creates a new resource and checks that the response contains an error message indicating that the user does not have the correct permissions.
func TestUserCanNotCreateDeniedResources(t *testing.T) {
	_, _, db, _, app, deregisterApp := setupTest(t)
	defer deregisterApp()

	// Create a resource without authentication
	responseWriter := th.MockWriter{}
	request, user, _ := th.NewRequest(
		http.MethodPost,
		`{"field": "test"}`,
		false,
		nil,
		nil,
	)

	assert.Nil(t, user, "User should be nil")

	t.Log("Creating a resource without authentication")
	app.ApiNew(db)(&responseWriter, request)

	data := responseWriter.GetWrittenData()
	assert.Contains(t, data, "no requested_by found in authorization header")
}

// TestUserCanUpdateAllowedResources tests that a user can update a resource if they have the correct permissions.
//
// It creates a resource for the logged-in user, and then checks that the response contains the updated record.
func TestUserCanUpdateAllowedResources(t *testing.T) {
	_, _, db, _, app, deregisterApp := setupTest(t)
	defer deregisterApp()

	// Create a resource for the logged-in user
	instance, user, userRollback := createTestResource(t, db, app, nil)
	defer userRollback()

	// Update the resource
	request, _, _ := th.NewRequest(
		http.MethodPut,
		`{"field": "updated_field"}`,
		true,
		user,
		map[string]string{"id": instance.GetIDString()},
	)

	response := executeApiCall(t, app.ApiUpdate(db), request)

	// Verify the update was successful
	var updatedInstance MockStruct
	json.Unmarshal([]byte(response), &updatedInstance)
	assert.Equal(t, instance.ID, updatedInstance.ID, "ID should be the same")
	assert.Equal(t, "updated_field", updatedInstance.Field, "Field should be the same")
	assert.Equal(t, user.ID, updatedInstance.CreatedByID, "CreatedBy should be the same")
	assert.Equal(t, user.ID, updatedInstance.UpdatedByID, "UpdatedBy should be the same")
}

// TestUserCanNotUpdateDeniedResources tests that a user can not update a resource if they don't have the correct permissions.
//
// It creates a resource for the logged-in user, and then checks that the response contains an error.
func TestUserCanNotUpdateDeniedResources(t *testing.T) {
	_, _, db, _, app, deregisterApp := setupTest(t)
	defer deregisterApp()

	// Create a resource for the logged-in user
	instance, _, userARollback := createTestResource(t, db, app, nil)
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

	response := executeApiCall(t, app.ApiUpdate(db), request)
	assert.Contains(t, response, "record not found", "The response should be an error")
}

// TestUserCanDeleteAllowedResources tests that a user can delete a resource if they have the correct permissions.
//
// It creates a resource for the logged-in user, and then checks that the response contains an error message indicating that the user does not have the correct permissions.
func TestUserCanDeleteAllowedResources(t *testing.T) {
	_, _, db, _, app, deregisterApp := setupTest(t)
	defer deregisterApp()

	// Create a resource for the logged-in user
	instance, user, userRollback := createTestResource(t, db, app, nil)
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
	executeApiCall(t, app.ApiDelete(db), deleteRequest)

	// Create a helper request to get the detail
	getRequest, _, _ := th.NewRequest(
		http.MethodGet,
		"",
		true,
		user,
		map[string]string{"id": instance.GetIDString()},
	)

	t.Log("Getting the detail for user")
	response := executeApiCall(
		t,
		app.ApiDetail(db),
		getRequest,
	)

	assert.Contains(t, response, "record not found", "The response should be an error")
}

// TestUserCanNotDeleteDeniedResources tests that a user cannot delete a resource if they don't have the correct permissions.
//
// It creates a resource for the logged-in user, and then checks that the response contains an error message indicating that the user does not have the correct permissions.
func TestUserCanNotDeleteDeniedResources(t *testing.T) {
	_, _, db, _, app, deregisterApp := setupTest(t)
	defer deregisterApp()

	// Create a resource for the logged-in user
	instance, userA, userARollback := createTestResource(t, db, app, nil)
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
	response := executeApiCall(t, app.ApiDelete(db), deleteRequest)
	assert.Contains(t, response, "record not found", "The response should be an error")

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
	responseB := executeApiCall(
		t,
		app.ApiDetail(db),
		getRequest,
	)

	var result MockStruct
	json.Unmarshal([]byte(responseB), &result)
	assert.Equal(t, instance.ID, result.ID, "ID should be the same")
	assert.Equal(t, instance.Field, result.Field, "Field should be the same")
	assert.Equal(t, instance.CreatedByID, result.CreatedByID, "CreatedBy should be the same")
}

// TestValidators tests that a user cannot create a resource with invalid
// values. It creates a new resource and checks that the response contains an
// error message indicating that the validation failed.
func TestCreateCallsValidators(t *testing.T) {
	_, _, db, _, app, deregisterApp := setupTest(t)
	defer deregisterApp()

	// Create a resource
	responseWriter := th.MockWriter{}
	request, _, rollback := th.NewRequest(
		http.MethodPost,
		`{"field": ""}`,
		true,
		nil,
		nil,
	)
	defer rollback()

	t.Log("Creating a resource")
	app.ApiNew(db)(&responseWriter, request)

	data := responseWriter.GetWrittenData()
	assert.Contains(t, data, "field cannot be empty", "The response should be an error")
}

func TestUpdateCallsValidators(t *testing.T) {
	_, _, db, _, app, deregisterApp := setupTest(t)
	defer deregisterApp()

	// Create a resource
	instance, user, rollback := createTestResource(t, db, app, nil)
	defer rollback()

	// Update the resource
	responseWriter := th.MockWriter{}
	request, _, _ := th.NewRequest(
		http.MethodPut,
		`{"field": ""}`,
		true,
		user,
		map[string]string{"id": instance.GetIDString()},
	)
	defer rollback()

	t.Log("Trying to update the resource")
	app.ApiUpdate(db)(&responseWriter, request)

	data := responseWriter.GetWrittenData()
	assert.Contains(t, data, "field cannot be empty", "The response should be an error")
}
