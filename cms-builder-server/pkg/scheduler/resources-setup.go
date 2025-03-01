package scheduler

import (
	"net/http"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	mgr "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
)

func SetupSchedulerTaskResource() *mgr.ResourceConfig {

	skipUserBinding := true // Scheduler Models don't have a created_by field

	permissions := server.RolePermissionMap{
		models.AdminRole:     server.AllAllowedAccess,
		models.SchedulerRole: server.AllAllowedAccess,
	}

	config := &mgr.ResourceConfig{
		Model:           SchedulerTask{},
		SkipUserBinding: skipUserBinding,
		Permissions:     permissions,
	}

	return config

}

func SetupSchedulerJobDefinitionResource(manager *mgr.ResourceManager, db *database.Database, jr JobRegistry) *mgr.ResourceConfig {

	skipUserBinding := true // Scheduler Models don't have a created_by field

	permissions := server.RolePermissionMap{
		models.AdminRole:     server.AllAllowedAccess,
		models.SchedulerRole: server.AllAllowedAccess,
	}

	routes := []server.Route{
		{
			Path:         "/job/run",
			Handler:      RunSchedulerTaskHandler(manager, db, jr),
			Name:         "job-run",
			RequiresAuth: true,
			Methods:      []string{http.MethodPost},
		},
	}

	config := &mgr.ResourceConfig{
		Model:           SchedulerJobDefinition{},
		SkipUserBinding: skipUserBinding,
		Permissions:     permissions,
		Routes:          routes,
	}

	return config

}
