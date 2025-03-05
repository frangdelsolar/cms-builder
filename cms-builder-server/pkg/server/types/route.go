package types

import "net/http"

type Route struct {
	Path         string           // route is the path for the route. i.e. /users/{id}
	Handler      http.HandlerFunc // handler is the handler for the route
	Name         string           // name is the name of the route
	RequiresAuth bool             // requiresAuth is a flag indicating if the route requires authentication
	Methods      []string         // method is the HTTP method for the route
}
