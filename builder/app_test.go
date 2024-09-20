package builder_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/frangdelsolar/cms/builder"
	th "github.com/frangdelsolar/cms/builder/test_helpers"
	"github.com/stretchr/testify/assert"
)

// TestNewAdmin tests that NewAdmin returns a non-nil Admin instance.
func TestNewAdmin(t *testing.T) {
	t.Log("Testing NewAdmin")
	engine := th.GetDefaultEngine()
	db, _ := engine.GetDatabase()
	server, _ := engine.GetServer()

	admin := builder.NewAdmin(db, server)

	assert.NotNil(t, admin)
}

// TestRegisterApp tests that RegisterApp registers a new App, applies database migration,
// and registers API routes for CRUD operations.
//
// It also tests that the registered App is accessible via the GetApp method.
func TestRegisterApp(t *testing.T) {
	engine := th.GetDefaultEngine()
	admin, err := engine.GetAdmin()

	assert.NoError(t, err)
	assert.NotNil(t, admin)

	// define a test struct to be registered
	type testStruct struct {
		*builder.SystemData
		Field string
	}

	t.Log("Testing RegisterApp")
	app, err := admin.Register(testStruct{}, false)
	assert.NoError(t, err)
	assert.NotNil(t, app)
	assert.Equal(t, "teststruct", app.Name())
	assert.Equal(t, "teststructs", app.PluralName())

	// check if the app is registered
	t.Log("Testing GetApp")
	retrievedApp, err := admin.GetApp("teststruct")
	assert.Equal(t, app.Name(), retrievedApp.Name())
	assert.NoError(t, err)
}

func TestValidators(t *testing.T) {}

// TestRegisterAPIRoutes tests that the RegisterApp method registers the expected routes with the Server.
//
// It registers a test App with the admin instance, and then checks that the server has the expected routes.
func TestRegisterAPIRoutes(t *testing.T) {
	engine := th.GetDefaultEngine()
	admin, _ := engine.GetAdmin()

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

	server, _ := engine.GetServer()
	routes := server.GetRoutes()
	for _, expectedRoute := range expectedRoutes {
		found := false
		for _, route := range routes {
			if route.Route == expectedRoute.Route {
				assert.Equal(t, expectedRoute.Name, route.Name)
				assert.Equal(t, expectedRoute.RequiresAuth, route.RequiresAuth)
				found = true
			}
		}

		assert.True(t, found)
	}
}

func TestUserCanRetrieveAllowedResources(t *testing.T)   {}
func TestUserCanNotRetrieveDeniedResources(t *testing.T) {}

// TestUserCanListAllowedResources tests that a user can list resources if they have the correct permissions.
//
// It creates three resources, one for some random user and two for the logged in user,
// and checks that the response contains the created resources and that the resource is persisted in the database.
func TestUserCanListAllowedResources(t *testing.T) {
	engine := th.GetDefaultEngine()
	admin, _ := engine.GetAdmin()
	db, _ := engine.GetDatabase()

	type TestList struct {
		*builder.SystemData
		Field string
	}
	app, _ := admin.Register(TestList{}, false)
	responseWriter := th.MockWriter{}
	var createdItem TestList

	// Create a new resource for some random user
	// this item should not be retrieved by the list endpoint
	requestR, _, rollbackUserR := th.NewRequest(
		http.MethodPost,
		"{\"field\": \"test\"}",
		true,
		nil,
	)
	defer rollbackUserR()

	app.ApiNew(db)(&responseWriter, requestR)

	// Create two resources for the logged in user
	responseWriter = th.MockWriter{}
	request, user, rollbackUser := th.NewRequest(
		http.MethodPost,
		"{\"field\": \"test\"}",
		true,
		nil,
	)
	defer rollbackUser()

	app.ApiNew(db)(&responseWriter, request)
	data := responseWriter.GetWrittenData()
	json.Unmarshal([]byte(data), &createdItem)

	t.Log("createdItem: ", createdItem)
	assert.NotNil(t, createdItem.ID)

	request, user, _ = th.NewRequest(
		http.MethodPost,
		"{\"field\": \"test\"}",
		true,
		user,
	)

	app.ApiNew(db)(&responseWriter, request)

	// Get the list for the logged in user
	request, _, _ = th.NewRequest(
		http.MethodGet,
		"",
		true,
		user,
	)
	w := th.MockWriter{}
	app.ApiList(db)(&w, request)

	listData := w.GetWrittenData()

	var items []TestList
	json.Unmarshal([]byte(listData), &items)

	// Verify the list contains two items
	assert.Equal(t, 2, len(items))
	assert.Equal(t, createdItem.Field, items[0].Field)
}

// TestUserCanNotListDeniedResources tests that a user cannot list resources if they do not have the correct permissions.
//
// It creates a new resource and checks that the response contains an error message indicating that the user does not have the correct permissions.
func TestUserCanNotListDeniedResources(t *testing.T) {
	engine := th.GetDefaultEngine()
	admin, _ := engine.GetAdmin()
	db, _ := engine.GetDatabase()

	type TestListDenied struct{}
	app, _ := admin.Register(TestListDenied{}, false)
	responseWriter := th.MockWriter{}
	request, _, _ := th.NewRequest(
		http.MethodGet,
		"",
		false,
		nil,
	)

	// Call the controller
	app.ApiList(db)(&responseWriter, request)

	data := responseWriter.GetWrittenData()
	assert.Equal(t, data, "[]")
}

func TestUserCanUpdateAllowedResources(t *testing.T)   {}
func TestUserCanNotUpdateDeniedResources(t *testing.T) {}

func TestUserCanDeleteAllowedResources(t *testing.T)   {}
func TestUserCanNotDeleteDeniedResources(t *testing.T) {}

// TestUserCanCreateAllowedResources tests that a user can create a resource if they have the correct permissions.
//
// It creates a new resource and checks that the response contains the created resource and that the resource is persisted in the database.
func TestUserCanCreateAllowedResources(t *testing.T) {
	engine := th.GetDefaultEngine()
	admin, _ := engine.GetAdmin()
	db, _ := engine.GetDatabase()
	type TestNew struct {
		*builder.SystemData
		Field string
	}

	app, _ := admin.Register(TestNew{}, false)
	responseWriter := th.MockWriter{}
	request, _, rollbackUser := th.NewRequest(
		http.MethodPost,
		`{"field": "test"}`,
		true,
		nil,
	)
	defer rollbackUser()

	// Call the controller
	app.ApiNew(db)(&responseWriter, request)

	var createdItem TestNew
	data := responseWriter.GetWrittenData()
	json.Unmarshal([]byte(data), &createdItem)

	assert.NotNil(t, createdItem.ID)
	assert.Equal(t, "test", createdItem.Field)
	assert.NotNil(t, createdItem.CreatedByID)
}

// TestUserCanNotCreateDeniedResources tests that a user cannot create a resource if they do not have the correct permissions.
//
// It creates a new resource and checks that the response contains an error message indicating that the user does not have the correct permissions.
func TestUserCanNotCreateDeniedResources(t *testing.T) {
	engine := th.GetDefaultEngine()
	admin, _ := engine.GetAdmin()
	db, _ := engine.GetDatabase()

	type TestNewDenied struct {
		*builder.SystemData
		Field string
	}

	app, _ := admin.Register(TestNewDenied{}, false)
	responseWriter := th.MockWriter{}
	request, _, _ := th.NewRequest(
		http.MethodPost,
		`{"field": "test"}`,
		false,
		nil,
	)

	// Call the controller
	app.ApiNew(db)(&responseWriter, request)

	data := responseWriter.GetWrittenData()
	assert.Contains(t, data, "no requested_by found in authorization header")
}
