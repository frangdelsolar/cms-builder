package server

import (
	"net/http"

	"github.com/NYTimes/gziphandler"
	authMiddlewares "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/middlewares"
	rlMiddlewares "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/request-logger/middlewares"
	svrHandlers "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server/handlers"
	svrMiddlewares "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server/middlewares"
	svrTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server/types"
)

func RunServer(s *svrTypes.Server, getRoutes svrTypes.GetRoutesFunc, certificates *svrTypes.TLSCertificateConfig) error {

	routes := getRoutes(s.BaseUrl)

	routes = append(routes,
		svrTypes.Route{
			Path:         "/",
			Handler:      svrHandlers.HealthCheck,
			Name:         "health",
			RequiresAuth: false,
			Methods:      []string{http.MethodGet},
		},
	)

	s.Logger.Info().Msgf("Initializing %d routes", len(routes))

	publicRouter := s.Root
	publicRouter.Use(
		svrMiddlewares.RecoveryMiddleware,
		authMiddlewares.AuthMiddleware(s.GodToken, s.GodUser, s.Firebase, s.DB, s.SystemUser),
		authMiddlewares.UserCookieMiddleware,
		rlMiddlewares.RequestLoggerMiddleware(s.DB),
		svrMiddlewares.LoggingMiddleware(s.LoggerConfig),
		svrMiddlewares.CorsMiddleware(s.AllowedOrigins),
		svrMiddlewares.TimeoutMiddleware,
		svrMiddlewares.RateLimitMiddleware(),
		gziphandler.GzipHandler,
	)

	routesSeen := map[string]bool{}

	for _, route := range routes {

		// Check for duplicates - We shouldn't have two routes with the same path
		if routesSeen[route.Path] {
			panic("Duplicate route: " + route.Path + " - " + route.Name)
		}
		routesSeen[route.Path] = true

		if route.RequiresAuth {
			continue
		}
		s.Logger.Warn().Str("path", route.Path).Msg("Public")

		methods := route.Methods
		methods = append(methods, http.MethodOptions)

		publicRouter.HandleFunc(route.Path, route.Handler).Name(route.Name).Methods(methods...)
	}

	authRouter := publicRouter.PathPrefix("/private").Subrouter()
	authRouter.Use(svrMiddlewares.ProtectedRouteMiddleware)

	for _, route := range routes {
		if !route.RequiresAuth {
			continue
		}

		methods := route.Methods
		methods = append(methods, http.MethodOptions)
		s.Logger.Error().Str("path", "/private"+route.Path).Msg("Private")

		authRouter.HandleFunc(route.Path, route.Handler).Name(route.Name).Methods(methods...)
	}

	s.Logger.Info().Msgf("Running server on port %s", s.Addr)

	if certificates == nil {
		return s.ListenAndServe()
	}

	return s.ListenAndServeTLS(certificates.CertFile, certificates.KeyFile)
}
