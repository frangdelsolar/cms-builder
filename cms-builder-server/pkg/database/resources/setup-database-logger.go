package resources

import (
	"net/http"

	auth "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/constants"
	authTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/types"
	dbHandlers "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/handlers"
	dbModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/models"
	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"
	loggerTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger/types"
	rmTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager/types"
	svrTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server/types"
)

func SetupDBLoggerResource(resourceManager *rmTypes.ResourceManager, db *dbTypes.DatabaseConnection, log *loggerTypes.Logger) *rmTypes.ResourceConfig {

	log.Info().Msg("Initializing Database Logger resource")

	skipUserBinding := true // DB Logs don't have a created_by field

	permissions := authTypes.RolePermissionMap{
		auth.AdminRole: []authTypes.CrudOperation{auth.OperationRead},
	}

	validators := rmTypes.ValidatorsMap{}
	handlers := &rmTypes.ApiHandlers{} // default
	routes := []svrTypes.Route{
		{
			Path:         "/api/database-timeline",
			Handler:      dbHandlers.TimelineHandler(resourceManager, db),
			Name:         "database-timeline",
			RequiresAuth: true,
			Methods:      []string{http.MethodGet},
		},
	}

	config := &rmTypes.ResourceConfig{
		Model:           dbModels.DatabaseLog{},
		SkipUserBinding: skipUserBinding,
		Validators:      validators,
		Permissions:     permissions,
		Handlers:        handlers,
		Routes:          routes,
	}

	return config
}
