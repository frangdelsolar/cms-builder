package utils

import (
	authTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/types"
	loggerTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger/types"
)

func UserIsAllowed(appPermissions authTypes.RolePermissionMap, userRoles []authTypes.Role, action authTypes.CrudOperation, resourceName string, log *loggerTypes.Logger) bool {

	// Loop over the user's roles and their associated permissions
	for _, role := range userRoles {
		if _, ok := appPermissions[role]; ok {
			for _, allowedAction := range appPermissions[role] {
				if allowedAction == action {
					log.Debug().Msgf("Granted access: User with role %s can %s resource %s", role, action, resourceName)
					return true
				}
			}
		}
	}
	log.Debug().Msgf("Denied access: User can not %s resource %s", action, resourceName)

	return false
}
