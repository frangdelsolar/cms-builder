package builder

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gertd/go-pluralize"
	"github.com/gorilla/mux"
)

var (
	ErrAdminNotInitialized = errors.New("admin not initialized")
)

type App interface{}

func GetAppName(a interface{}) string {
	modelName := fmt.Sprintf("%T", a)
	name := modelName[strings.LastIndex(modelName, ".")+1:]
	name = strings.ToLower(name)
	return name
}

func Pluralize(word string) string {
	p := pluralize.NewClient()
	return p.Plural(word)
}

type Admin struct {
	Apps   []App
	db     *Database
	server *Server
}

type AdminConfig struct {
}

func NewAdmin(config *AdminConfig, db *Database, server *Server) *Admin {
	return &Admin{
		Apps:   make([]App, 0),
		db:     db,
		server: server,
	}
}

func (a *Admin) Register(app App) {
	log.Debug().Interface("Register", app).Msg("Registering app")

	a.Apps = append(a.Apps, app)

	// Apply database migration
	a.db.Migrate(app)

	// Register api routes
	appName := GetAppName(app)
	a.RegisterApp(Pluralize(appName), app)
}

func (a *Admin) RegisterApp(appName string, app interface{}) {
	appRoutes := a.server.root.PathPrefix("/api/" + appName).Subrouter()
	appRoutes.HandleFunc("", func(w http.ResponseWriter, r *http.Request) {
		List(app, a.db, w, r)
	})
	appRoutes.HandleFunc("/new", func(w http.ResponseWriter, r *http.Request) {
		New(app, a.db, w, r)
	})
	appRoutes.HandleFunc("/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		Get(id, app, a.db, w, r)
	})
	appRoutes.HandleFunc("/{id}/delete", func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		Delete(id, app, a.db, w, r)
	})
	appRoutes.HandleFunc("/{id}/edit", func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		Update(id, app, a.db, w, r)
	})
}
