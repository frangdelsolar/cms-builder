package builder

import (
	"fmt"
	"strings"

	"github.com/gertd/go-pluralize"
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
	router *Router
}

type AdminConfig struct {
}

func NewAdmin(config *AdminConfig, db *Database, server *Server) *Admin {
	return &Admin{
		Apps:   make([]App, 0),
		db:     db,
		router: server.Router,
	}
}

func (a *Admin) Register(app App) {
	log.Debug().Interface("Register", app).Msg("Registering app")
	a.Apps = append(a.Apps, app)
	// Apply database migration
	a.db.Migrate(app)

	// Register api routes
	appName := GetAppName(app)
	a.router.RegisterApp(Pluralize(appName), app)

}
