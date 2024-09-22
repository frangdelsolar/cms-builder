package builder_test

import (
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

// FieldValidator validates the given field.
//
// It checks if the field value is empty. If the field value is empty, it returns an error with the message
// "{field name} cannot be empty". Otherwise, it returns nil.
//
// Parameters:
// - fieldName: the name of the field to be validated.
// - instance: a map[string]interface{} representing the instance to be validated.
//
// Returns:
// - error: an error if the field value is empty, otherwise nil.
func FieldValidator(fieldName string, instance map[string]interface{}, output *builder.FieldValidationError) *builder.FieldValidationError {
	fieldValue := fmt.Sprint(instance[fieldName])
	if fieldValue == "" {
		output.Error = fieldName + " cannot be empty"
	}

	return output
}

// setupTest sets up a default Builder instance, Admin, Database, and Server instances,
// and registers a new App with a MockStruct type. It also sets up a field validator
// for the "field" key on the MockStruct type. The function returns the instances
// and a callback to be used to deregister the App after the test is finished.
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

	data := "{\"field\": \"" + th.RandomString(10) + "\"}"
	request, user, rollback := th.NewRequest(http.MethodPost, data, true, user, nil)

	t.Logf("Creating new resource for user: %v", user.ID)

	var createdItem MockStruct
	response, err := executeApiCall(t, app.ApiNew(db), request, &createdItem)

	assert.NoError(t, err, "ApiNew should not return an error")
	assert.True(t, response.Success, "ApiNew should return a success response")
	return &createdItem, user, rollback
}

// executeApiCall executes the given API call handler function with the given request and stores the response in the given value.
//
// It logs the request method and body, creates a new MockWriter, and calls the API call handler function with the MockWriter and the request. It then parses the response from the MockWriter and stores it in the given value. Finally, it asserts that the parsing did not return an error.
//
// Parameters:
// - t: the testing.T instance.
// - apiCall: the API call handler function to be executed.
// - request: the request to be passed to the API call handler function.
// - v: the value to store the response in.
//
// Returns:
// - builder.Response: the parsed response from the API call handler function.
func executeApiCall(t *testing.T, apiCall builder.HandlerFunc, request *http.Request, v interface{}) (builder.Response, error) {
	t.Log("Executing API call", request.Method, request.Body)
	writer := th.MockWriter{}
	apiCall(&writer, request)

	return builder.ParseResponse(writer.Buffer.Bytes(), v)
}

// TestNewAdmin tests that NewAdmin returns a non-nil Admin instance.
func TestNewAdmin(t *testing.T) {
	t.Log("Testing NewAdmin")
	engine := th.GetDefaultEngine()
	db, _ := engine.GetDatabase()
	server, _ := engine.GetServer()

	admin := builder.NewAdmin(db, server, engine)

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
	var result MockStruct
	response, err := executeApiCall(
		t,
		app.ApiDetail(db),
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

	responseA, err := executeApiCall(t, app.ApiDetail(db), requestA, nil)
	assert.Error(t, err, "ApiDetail should return an error")
	assert.False(t, responseA.Success, "ParseResponse should return an error response")
	assert.Contains(t, responseA.Message, "Failed to get", "The response should be an error")

	// User B tries to get the detail for user A
	requestB, _, _ := th.NewRequest(
		http.MethodGet,
		"",
		true,
		userB,
		map[string]string{"id": instanceA.GetIDString()},
	)

	responseB, err := executeApiCall(t, app.ApiDetail(db), requestB, nil)
	assert.Error(t, err, "ApiDetail should return an error")
	assert.False(t, responseB.Success, "ParseResponse should return an error response")
	assert.Contains(t, responseB.Message, "Failed to get", "The response should be an error")
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
	var result []MockStruct
	response, err := executeApiCall(
		t,
		app.ApiList(db),
		request,
		&result,
	)

	assert.NoError(t, err, "ApiList should not return an error")
	assert.NotNil(t, response, "ApiList should return a non-nil response")
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
	var result []MockStruct
	response, err := executeApiCall(
		t,
		app.ApiList(db),
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
	request, user, _ := th.NewRequest(
		http.MethodPost,
		`{"field": "test"}`,
		false,
		nil,
		nil,
	)

	assert.Nil(t, user, "User should be nil")

	t.Log("Creating a resource without authentication")
	var result []MockStruct
	response, err := executeApiCall(
		t,
		app.ApiNew(db),
		request,
		&result,
	)
	assert.Error(t, err, "ApiNew should return an error")
	assert.NotNil(t, response, "ApiNew should return a non-nil response")
	assert.Contains(t, response.Message, "user not authenticated", "The response should be an error message")
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

	var updatedInstance MockStruct
	response, err := executeApiCall(t, app.ApiUpdate(db), request, &updatedInstance)

	assert.NoError(t, err, "ApiUpdate should not return an error")
	assert.NotNil(t, response, "ApiUpdate should return a non-nil response")

	// Verify the update was successful
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

	response, err := executeApiCall(t, app.ApiUpdate(db), request, nil)

	assert.Error(t, err, "ApiUpdate should return an error")
	assert.NotNil(t, response, "ApiUpdate should return a non-nil response")
	assert.Contains(t, response.Message, "record not found", "The response should be an error")
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
	var resultA MockStruct
	responseA, err := executeApiCall(t, app.ApiDelete(db), deleteRequest, &resultA)

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
	var resultB MockStruct
	responseB, err := executeApiCall(
		t,
		app.ApiDetail(db),
		getRequest,
		&resultB,
	)

	assert.NoError(t, err, "ApiDetail should not return an error")
	assert.NotNil(t, responseB, "ApiDetail should return a non-nil response")
	assert.Equal(t, responseB.Success, false, "Success should be false")
	assert.Contains(t, responseB.Message, "Failed to get", "The response should be an error")
	assert.Equal(t, resultB, MockStruct{}, "The result should be empty")
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
	var resultA MockStruct
	responseA, err := executeApiCall(t, app.ApiDelete(db), deleteRequest, &resultA)

	assert.NoError(t, err, "ApiDetail should not return an error")
	assert.NotNil(t, responseA, "ApiDetail should return a non-nil response")
	assert.Equal(t, responseA.Success, false, "Success should be false")
	assert.Contains(t, responseA.Message, "record not found", "The response should be an error")
	assert.Equal(t, resultA, MockStruct{}, "The result should be empty")

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
	var resultB MockStruct
	responseB, err := executeApiCall(
		t,
		app.ApiDetail(db),
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
	_, _, db, _, app, deregisterApp := setupTest(t)
	defer deregisterApp()

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
	var result MockStruct
	response, err := executeApiCall(
		t,
		app.ApiNew(db),
		request,
		&result,
	)

	assert.NoError(t, err, "executeApiCall should not return an error")
	assert.False(t, response.Success, "ApiDetail should return an error")
	assert.Contains(t, response.Message, "Validation failed", "The response should be an error")
}

// TestUpdateCallsValidators tests that a user cannot update a resource with
// invalid values. It creates a resource, tries to update it with invalid values,
// and checks that the response contains an error message indicating that the
// validation failed.
func TestUpdateCallsValidators(t *testing.T) {
	_, _, db, _, app, deregisterApp := setupTest(t)
	defer deregisterApp()

	// Create a resource
	instance, user, rollback := createTestResource(t, db, app, nil)
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
	var result MockStruct
	response, _ := executeApiCall(
		t,
		app.ApiUpdate(db),
		request,
		&result,
	)

	// assert.NoError(t, err, "executeApiCall should not return an error")
	assert.False(t, response.Success, "ApiDetail should return an error")
	assert.Contains(t, response.Message, "Validation failed", "The response should be an error")
}

// TestUserCanNotReplaceCreatedByIDOnCreate tests that a user cannot create a resource with a createdById or updatedById that is not their own user ID.
func TestUserCanNotReplaceCreatedByIDOnCreate(t *testing.T) {
	_, _, db, _, app, deregisterApp := setupTest(t)
	defer deregisterApp()

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

	var instance MockStruct
	response, err := executeApiCall(t, app.ApiNew(db), request, &instance)
	assert.NoError(t, err, "executeApiCall should not return an error")
	assert.NotNil(t, response, "ApiNew should return a non-nil response")

	assert.Equal(t, user.GetIDString(), fmt.Sprintf("%d", instance.CreatedByID), "CreatedByID should be the logged in user")
	assert.Equal(t, user.GetIDString(), fmt.Sprintf("%d", instance.UpdatedByID), "UpdatedByID should be the logged in user")
}

// TestUserCanNotReplaceCreatedByIDOnUpdate tests that a user cannot update a resource with a createdById or updatedById that is not their own user ID.
func TestUserCanNotReplaceCreatedByIDOnUpdate(t *testing.T) {
	_, _, db, _, app, deregisterApp := setupTest(t)
	defer deregisterApp()

	// Create a resource for the logged-in user
	instance, user, userRollback := createTestResource(t, db, app, nil)
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
	var updatedInstance MockStruct
	response, _ := executeApiCall(t, app.ApiUpdate(db), request, &updatedInstance)

	assert.NotNil(t, response, "ApiUpdate should return a non-nil response")
	assert.Equal(t, true, response.Success, "ApiUpdate should return a success response")

	// Verify the update was successful
	assert.Equal(t, instance.ID, updatedInstance.ID, "ID should be the same")
	assert.Equal(t, user.GetIDString(), fmt.Sprintf("%d", updatedInstance.CreatedByID), "CreatedByID should be the logged in user")
	assert.Equal(t, user.GetIDString(), fmt.Sprintf("%d", updatedInstance.UpdatedByID), "UpdatedByID should be the logged in user")
}

// TestUserCanNotReplaceInstanceIDOnUpdate tests that a user cannot update a resource with an instance ID that is not the same as the one in the request.
func TestUserCanNotReplaceInstanceIDOnUpdate(t *testing.T) {
	_, _, db, _, app, deregisterApp := setupTest(t)
	defer deregisterApp()

	// Create a resource for the logged-in user
	instance, user, userRollback := createTestResource(t, db, app, nil)
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
	var updatedInstance MockStruct
	response, _ := executeApiCall(t, app.ApiUpdate(db), request, &updatedInstance)
	assert.NotNil(t, response, "ApiUpdate should return a non-nil response")
	assert.Equal(t, true, response.Success, "ApiUpdate should return a success response")
	// Verify the update was successful
	assert.Equal(t, instance.GetIDString(), updatedInstance.GetIDString(), "ID should remain the same")
	assert.Equal(t, user.GetIDString(), fmt.Sprintf("%d", updatedInstance.CreatedByID), "CreatedByID should be the logged in user")
	assert.Equal(t, user.GetIDString(), fmt.Sprintf("%d", updatedInstance.UpdatedByID), "UpdatedByID should be the logged in user")
}
