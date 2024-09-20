package builder_test

import (
	"net/http"
	"testing"

	"github.com/frangdelsolar/cms/builder"
	th "github.com/frangdelsolar/cms/builder/test_helpers"
	"github.com/stretchr/testify/assert"
)

func TestNewAdmin(t *testing.T) {
	t.Log("Testing NewAdmin")
	engine := th.GetDefaultEngine()
	db, _ := engine.GetDatabase()
	server, _ := engine.GetServer()

	admin := builder.NewAdmin(db, server)

	assert.NotNil(t, admin)
}

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
