package scheduler

import (
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	manager "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
)

func SetupSchedulerTaskResource() *manager.ResourceConfig {

	skipUserBinding := true // Scheduler Models don't have a created_by field

	permissions := server.RolePermissionMap{
		models.AdminRole:     []server.CrudOperation{server.OperationRead},
		models.SchedulerRole: server.AllAllowedAccess,
	}

	config := &manager.ResourceConfig{
		Model:           models.SchedulerTask{},
		SkipUserBinding: skipUserBinding,
		Permissions:     permissions,
	}

	return config

}

func SetupSchedulerJobDefinitionResource() *manager.ResourceConfig {

	skipUserBinding := true // Scheduler Models don't have a created_by field

	permissions := server.RolePermissionMap{
		models.AdminRole:     []server.CrudOperation{server.OperationRead},
		models.SchedulerRole: server.AllAllowedAccess,
	}

	config := &manager.ResourceConfig{
		Model:           models.SchedulerJobDefinition{},
		SkipUserBinding: skipUserBinding,
		Permissions:     permissions,
	}

	return config

}
