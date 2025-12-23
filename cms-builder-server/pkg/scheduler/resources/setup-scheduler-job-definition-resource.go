package resources

import (
	"net/http"

	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"

	authConstants "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/constants"
	authTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/types"
	rmPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"
	rmTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager/types"
	schHandlers "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/scheduler/handlers"
	schModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/scheduler/models"
	schTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/scheduler/types"
	svrTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server/types"
)

func SetupSchedulerJobDefinitionResource(manager *rmPkg.ResourceManager, db *dbTypes.DatabaseConnection, jr schTypes.JobRegistry, runScheduler bool) *rmTypes.ResourceConfig {

	skipUserBinding := true // Scheduler Models don't have a created_by field

	permissions := authTypes.RolePermissionMap{
		authConstants.AdminRole:     authConstants.AllAllowedAccess,
		authConstants.SchedulerRole: authConstants.AllAllowedAccess,
	}

	routes := []svrTypes.Route{
		{
			Path:         "/job/run",
			Handler:      schHandlers.RunSchedulerTaskHandler(manager, db, jr, runScheduler),
			Name:         "job-run",
			RequiresAuth: true,
			Methods:      []string{http.MethodPost},
		},
	}

	config := &rmTypes.ResourceConfig{
		Model:           schModels.SchedulerJobDefinition{},
		SkipUserBinding: skipUserBinding,
		Permissions:     permissions,
		Routes:          routes,
	}

	return config

}
