package builder

import (
	"errors"
	"fmt"
)

var (
	ErrAdminNotInitialized = errors.New("admin not initialized")
)

type Admin struct {
	apps   map[string]App
	db     *Database
	server *Server
}

// NewAdmin creates a new instance of the Admin, which is a central
// configuration and management structure for managing applications.
//
// Parameters:
// - db: A pointer to the Database instance to use for database operations.
// - server: A pointer to the Server instance to use for registering API routes.
//
// Returns:
// - *Admin: A pointer to the new Admin instance.
func NewAdmin(db *Database, server *Server) *Admin {
	return &Admin{
		apps:   make(map[string]App),
		db:     db,
		server: server,
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
func (a *Admin) GetApp(appName string) (App, error) {
	if app, ok := a.apps[appName]; ok {
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
func (a *Admin) Register(model interface{}, skipUserBinding bool) App {
	log.Debug().Interface("App", model).Msg("Registering app")

	app := App{
		model:           model,
		skipUserBinding: skipUserBinding,
		admin:           a,
		validators:      make(map[string]FieldValidationFunc),
	}

	a.apps[app.Name()] = app
	a.db.Migrate(app.model)

	a.registerAPIRoutes(app)

	return app
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
func (a *Admin) registerAPIRoutes(app App) {
	baseRoute := "/api/" + app.PluralName()

	a.server.AddRoute(
		baseRoute,
		app.apiList(a.db),
		app.Name()+"-list",
		true,
	)

	a.server.AddRoute(
		baseRoute+"/new",
		app.apiNew(a.db),
		app.Name()+"-new",
		true,
	)

	a.server.AddRoute(
		baseRoute+"/{id}",
		app.apiDetail(a.db),
		app.Name()+"-get",
		true,
	)

	a.server.AddRoute(
		baseRoute+"/{id}/delete",
		app.apiDelete(a.db),
		app.Name()+"-delete",
		true,
	)

	a.server.AddRoute(
		baseRoute+"/{id}/update",
		app.apiUpdate(a.db),
		app.Name()+"-update",
		true,
	)
}
