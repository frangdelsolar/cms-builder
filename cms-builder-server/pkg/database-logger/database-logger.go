package databaselogger

import (
	"net/http"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	manager "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
)

func SetupDBLoggerResource(resourceManager *manager.ResourceManager, db *database.Database, log *logger.Logger) *manager.ResourceConfig {

	log.Info().Msg("Initializing Database Logger resource")

	skipUserBinding := true // DB Logs don't have a created_by field

	permissions := server.RolePermissionMap{
		models.AdminRole: []server.CrudOperation{server.OperationRead},
	}

	validators := manager.ValidatorsMap{}
	handlers := &manager.ApiHandlers{} // default
	routes := []server.Route{
		{
			Path:         "/database-logs/timeline",
			Handler:      TimelineHandler(resourceManager, db),
			Name:         "database-logs:timeline",
			RequiresAuth: true,
			Methods:      []string{http.MethodGet},
		},
	}

	config := &manager.ResourceConfig{
		Model:           models.DatabaseLog{},
		SkipUserBinding: skipUserBinding,
		Validators:      validators,
		Permissions:     permissions,
		Handlers:        handlers,
		Routes:          routes,
	}

	return config
}
