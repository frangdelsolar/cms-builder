package resources

import (
	authConstants "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/constants"
	authTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/types"
	rmTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager/types"
	schModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/scheduler/models"
)

func SetupSchedulerTaskResource() *rmTypes.ResourceConfig {

	skipUserBinding := true // Scheduler Models don't have a created_by field

	permissions := authTypes.RolePermissionMap{
		authConstants.AdminRole:     authConstants.AllAllowedAccess,
		authConstants.SchedulerRole: authConstants.AllAllowedAccess,
	}

	config := &rmTypes.ResourceConfig{
		Model:           schModels.SchedulerTask{},
		SkipUserBinding: skipUserBinding,
		Permissions:     permissions,
	}

	return config

}
