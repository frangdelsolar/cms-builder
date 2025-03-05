package requestlogger

import (
	"net/http"

	authConstants "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/constants"
	authTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/types"
	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"
	loggerTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger/types"
	rmHandlers "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/request-logger/handlers"
	rmModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/request-logger/models"
	rmTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager/types"
	svrTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server/types"
)

func SetupRequestLoggerResource(resourceManager *rmTypes.ResourceManager, db *dbTypes.DatabaseConnection, log *loggerTypes.Logger) *rmTypes.ResourceConfig {

	log.Info().Msg("Initializing Request Logger resource")

	skipUserBinding := true // Request Logs don't have a created_by field

	permissions := authTypes.RolePermissionMap{
		authConstants.AdminRole: []authTypes.CrudOperation{authConstants.OperationRead},
	}

	validators := rmTypes.ValidatorsMap{}
	handlers := &rmTypes.ApiHandlers{} // default
	routes := []svrTypes.Route{
		{
			Path:         "/request-logs-stats",
			Handler:      rmHandlers.RequestStatsHandler(resourceManager, db),
			Name:         "request-logs-stats",
			RequiresAuth: true,
			Methods:      []string{http.MethodGet},
		},
		{
			Path:         "/request-logs/{id}", // TODO modify to make it ?traceId
			Handler:      rmHandlers.RequestLogHandler(resourceManager, db),
			Name:         "request-logs-detail",
			RequiresAuth: true,
			Methods:      []string{http.MethodGet},
		},
	}

	config := &rmTypes.ResourceConfig{
		Model:           rmModels.RequestLog{},
		SkipUserBinding: skipUserBinding,
		Validators:      validators,
		Permissions:     permissions,
		Handlers:        handlers,
		Routes:          routes,
	}

	return config
}
