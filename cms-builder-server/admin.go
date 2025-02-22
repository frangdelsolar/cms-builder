package builder

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/invopop/jsonschema"
)

var (
	ErrAdminNotInitialized = errors.New("admin not initialized")
)

type Admin struct {
	apps    map[string]App
	Builder *Builder
}

// NewAdmin creates a new instance of the Admin, which is a central
// configuration and management structure for managing applications.
//
// Parameters:
// - builder: A pointer to the Builder instance to use for building App instances.
//
// Returns:
// - *Admin: A pointer to the new Admin instance.
func NewAdmin(builder *Builder) *Admin {
	return &Admin{
		apps:    make(map[string]App),
		Builder: builder,
	}
}

// GetApp returns the App instance associated with the given name.
//
// Parameters:
// - appName: The name of the App to retrieve.
//
// Returns:
// - App: The App instance associated with the given name if found.
// - error: An error if the App is not found.

// TODO: Fix so that accepts the interface and then looks for the name just like the model would do.... i.e. User{} -> user
func (a *Admin) GetApp(appName string) (App, error) {
	lowerAppName := strings.ToLower(appName)
	if app, ok := a.apps[lowerAppName]; ok {
		return app, nil
	}

	return App{}, fmt.Errorf("app not found: %s", appName)
}

// Register adds a new App to the Admin instance, applies database migration, and
// registers API routes for CRUD operations.
//
// Parameters:
// - model: The model to register.
// - skipUserBinding: Whether to skip user binding which is used for filtering db queries by userId
func (a *Admin) Register(model interface{}, skipUserBinding bool, permissions RolePermissionMap, api *ApiHandlers) (*App, error) {

	if api == nil {
		api = &ApiHandlers{
			Create: DefaultCreateHandler,
			Delete: DefaultDeleteHandler,
			Detail: DefaultDetailHandler,
			List:   DefaultListHandler,
			Update: DefaultUpdateHandler,
		}
	}

	app := &App{
		Model:           model,
		SkipUserBinding: skipUserBinding,
		Admin:           a,
		Validators:      make(ValidatorsMap),
		Permissions:     permissions,
		Api:             api,
	}

	// check the app is not already registered
	_, err := a.GetApp(app.Name())
	if err == nil {
		// If app isn't found it will return an error, which means it doesn't exist
		// In other words. We are expecting an error here. Error means slot is free for the new app
		return &App{}, fmt.Errorf("app already registered: %s", app.Name())
	}

	// register the app
	appName := strings.ToLower(app.Name())
	a.apps[appName] = *app

	// apply migrations
	a.Builder.DB.Migrate(app.Model)

	// register routes
	a.registerAPIRoutes(*app)

	return app, nil
}

// Unregister removes the given app from the Admin instance.
//
// If the app is not found, it returns an error.
//
// Parameters:
// - appName: The name of the app to be unregistered.
//
// Returns:
// - error: An error if the app is not found.
func (a *Admin) Unregister(appName string) error {

	lowerAppName := strings.ToLower(appName)
	if _, ok := a.apps[lowerAppName]; !ok {
		return fmt.Errorf("app not found: %s", appName)
	}

	delete(a.apps, lowerAppName)
	return nil
}

// registerAPIRoutes registers API routes for the given App.
//
// It takes two arguments:
//   - appName: The plural name of the App, which is used to generate the base route.
//   - app: The App struct to use for generating the API routes.
//
// It registers the following API routes:
//   - GET /{appName}: Returns a list of all App instances.
//   - POST /{appName}/new: Creates a new App instance.
//   - GET /{appName}/{id}: Returns the App instance with the given ID.
//   - DELETE /{appName}/{id}/delete: Deletes the App instance with the given ID.
//   - PUT /{appName}/{id}/update: Updates the App instance with the given ID.
//
// All CRUD routes are protected by authentication middleware.
func (a *Admin) registerAPIRoutes(app App) {

	if app.Api == (&ApiHandlers{}) ||
		app.Api == nil ||
		app.Api.List == nil ||
		app.Api.Detail == nil ||
		app.Api.Create == nil ||
		app.Api.Update == nil ||
		app.Api.Delete == nil {

		log.Error().Msg("API routes not registered for app: " + app.Name())
		return
	}

	kebabName := app.KebabPluralName()

	baseRoute := "/api/" + kebabName + "/"
	protectedRoute := true

	a.Builder.Server.AddRoute(
		baseRoute+"schema/",
		func(w http.ResponseWriter, r *http.Request) {
			schema := jsonschema.Reflect(app.Model)
			SendJsonResponse(w, http.StatusOK, schema, fmt.Sprintf("Schema for %s", app.Name()))
		},
		kebabName+"-schema",
		!protectedRoute,
		http.MethodGet,
		nil,
	)

	a.Builder.Server.AddRoute(
		baseRoute,
		app.ApiList(a.Builder.DB),
		kebabName+"-list",
		protectedRoute,
		http.MethodGet,
		nil,
	)

	a.Builder.Server.AddRoute(
		baseRoute+"new/",
		app.ApiCreate(a.Builder.DB),
		kebabName+"-new",
		protectedRoute,
		http.MethodPost,
		app.Model,
	)

	a.Builder.Server.AddRoute(
		baseRoute+"{id}/",
		app.ApiDetail(a.Builder.DB),
		kebabName+"-get",
		protectedRoute,
		http.MethodGet,
		nil,
	)

	a.Builder.Server.AddRoute(
		baseRoute+"{id}/delete/",
		app.ApiDelete(a.Builder.DB),
		kebabName+"-delete",
		protectedRoute,
		http.MethodDelete,
		nil,
	)

	a.Builder.Server.AddRoute(
		baseRoute+"{id}/update/",
		app.ApiUpdate(a.Builder.DB),
		kebabName+"-update",
		protectedRoute,
		http.MethodPut,
		app.Model,
	)
}

// AddApiRoute adds an endpoint that returns a JSON response with information about
// all the registered apps, including their name, plural name, snake case name,
// kebab case name, endpoints, and their respective HTTP methods.
//
// The structure of the response is as follows:
//
// [
//
//	{
//	  "name": "string",
//	  "pluralName": "string",
//	  "snakeName": "string",
//	  "kebabName": "string",
//	  "snakePluralName": "string",
//	  "kebabPluralName": "string",
//	  "endpoints": {
//	    "schema": {
//	      "method": "string",
//	      "path": "string",
//	    },
//	    // ... other endpoints ...
//	  }
//	},
//	// ... other apps ...
//
// ]
func (a *Admin) AddApiRoute() {

	s := a.Builder.Server
	s.AddRoute(
		"/api",
		func(w http.ResponseWriter, r *http.Request) {
			type Endpoint struct {
				Method string `json:"method"`
				Path   string `json:"path"`
			}
			type appInfo struct {
				Name        string              `json:"name"`
				Plural      string              `json:"pluralName"`
				Snake       string              `json:"snakeName"`
				Kebab       string              `json:"kebabName"`
				SnakePlural string              `json:"snakePluralName"`
				KebabPlural string              `json:"kebabPluralName"`
				Endpoints   map[string]Endpoint `json:"endpoints"`
			}
			output := make([]appInfo, 0)

			for _, app := range s.Builder.Admin.apps {
				baseUrl := config.GetString(EnvKeys.BaseUrl) + "/api/" + app.KebabPluralName()

				data := appInfo{
					Name:        app.Name(),
					Plural:      app.PluralName(),
					Snake:       app.SnakeName(),
					Kebab:       app.KebabName(),
					SnakePlural: app.SnakePluralName(),
					KebabPlural: app.KebabPluralName(),
					Endpoints: map[string]Endpoint{
						"schema": {
							Method: http.MethodGet,
							Path:   baseUrl + "/schema",
						},
					},
				}

				output = append(output, data)
			}

			SendJsonResponse(w, http.StatusOK, output, "ok")
		},
		"endpoints",
		false,
		http.MethodGet,
		nil,
	)
}
