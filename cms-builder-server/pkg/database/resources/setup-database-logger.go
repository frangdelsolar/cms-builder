package resources

import (
	"net/http"

	dbHandlers "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/handlers"
	dbModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/models"
	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"
	loggerTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger/types"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	manager "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
)

func SetupDBLoggerResource(resourceManager *manager.ResourceManager, db *dbTypes.DatabaseConnection, log *loggerTypes.Logger) *manager.ResourceConfig {

	log.Info().Msg("Initializing Database Logger resource")

	skipUserBinding := true // DB Logs don't have a created_by field

	permissions := server.RolePermissionMap{
		models.AdminRole: []server.CrudOperation{server.OperationRead},
	}

	validators := manager.ValidatorsMap{}
	handlers := &manager.ApiHandlers{} // default
	routes := []server.Route{
		{
			Path:         "/api/database-timeline",
			Handler:      dbHandlers.TimelineHandler(resourceManager, db),
			Name:         "database-timeline",
			RequiresAuth: true,
			Methods:      []string{http.MethodGet},
		},
	}

	config := &manager.ResourceConfig{
		Model:           dbModels.DatabaseLog{},
		SkipUserBinding: skipUserBinding,
		Validators:      validators,
		Permissions:     permissions,
		Handlers:        handlers,
		Routes:          routes,
	}

	return config
}
