package auth

// RolePermissionMap maps roles to their allowed CRUD operations.
type RolePermissionMap map[Role][]CrudOperation

// HasPermission checks if a user with the given roles has permission to perform the specified action.
func (p RolePermissionMap) HasPermission(userRoles []Role, action CrudOperation) bool {
	for _, role := range userRoles {
		if allowedActions, ok := p[role]; ok {
			for _, allowedAction := range allowedActions {
				if allowedAction == action {
					return true
				}
			}
		}
	}
	return false
}
