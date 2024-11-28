package builder

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrAdminNotInitialized = errors.New("admin not initialized")
)

type Admin struct {
	apps    map[string]App
	db      *Database
	server  *Server
	builder *Builder
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
func NewAdmin(db *Database, server *Server, builder *Builder) *Admin {
	return &Admin{
		apps:    make(map[string]App),
		db:      db,
		server:  server,
		builder: builder,
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
func (a *Admin) Register(model interface{}, skipUserBinding bool) (App, error) {

	app := App{
		model:           model,
		skipUserBinding: skipUserBinding,
		admin:           a,
		validators:      make(ValidatorsMap),
	}

	// check the app is not already registered
	_, err := a.GetApp(app.Name())
	if err == nil {
		// If app isn't found it will return an error, which means it doesn't exist
		// In other words. We are expecting an error here. Error means slot is free for the new app
		return App{}, fmt.Errorf("app already registered: %s", app.Name())
	}

	// register the app
	a.apps[app.Name()] = app

	// apply migrations
	a.db.Migrate(app.model)

	// register CRUD routes
	a.registerAPIRoutes(app)

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
	baseRoute := "/api/" + app.PluralName()
	protectedRoute := true

	a.server.AddRoute(
		baseRoute,
		app.ApiList(a.db),
		app.Name()+"-list",
		protectedRoute,
	)

	a.server.AddRoute(
		baseRoute+"/new",
		app.ApiNew(a.db),
		app.Name()+"-new",
		protectedRoute,
	)

	a.server.AddRoute(
		baseRoute+"/{id}",
		app.ApiDetail(a.db),
		app.Name()+"-get",
		protectedRoute,
	)

	a.server.AddRoute(
		baseRoute+"/{id}/delete",
		app.ApiDelete(a.db),
		app.Name()+"-delete",
		protectedRoute,
	)

	a.server.AddRoute(
		baseRoute+"/{id}/update",
		app.ApiUpdate(a.db),
		app.Name()+"-update",
		protectedRoute,
	)
}
