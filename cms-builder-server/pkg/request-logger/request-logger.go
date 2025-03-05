package requestlogger

import (
	"net/http"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	loggerTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger/types"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	manager "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
)

func SetupRequestLoggerResource(resourceManager *manager.ResourceManager, db *database.Database, log *loggerTypes.Logger) *manager.ResourceConfig {

	log.Info().Msg("Initializing Request Logger resource")

	skipUserBinding := true // Request Logs don't have a created_by field

	permissions := server.RolePermissionMap{
		models.AdminRole: []server.CrudOperation{server.OperationRead},
	}

	validators := manager.ValidatorsMap{}
	handlers := &manager.ApiHandlers{} // default
	routes := []server.Route{
		{
			Path:         "/request-logs-stats",
			Handler:      RequestStatsHandler(resourceManager, db),
			Name:         "request-logs-stats",
			RequiresAuth: true,
			Methods:      []string{http.MethodGet},
		},
		{
			Path:         "/request-logs/{id}", // TODO modify to make it ?traceId
			Handler:      RequestLogHandler(resourceManager, db),
			Name:         "request-logs-detail",
			RequiresAuth: true,
			Methods:      []string{http.MethodGet},
		},
	}

	config := &manager.ResourceConfig{
		Model:           models.RequestLog{},
		SkipUserBinding: skipUserBinding,
		Validators:      validators,
		Permissions:     permissions,
		Handlers:        handlers,
		Routes:          routes,
	}

	return config
}
