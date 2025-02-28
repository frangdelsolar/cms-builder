package server

import (
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
)

// CrudOperation represents the types of CRUD operations.
type CrudOperation string

// Predefined CRUD operations.
const (
	OperationCreate CrudOperation = "create"
	OperationDelete CrudOperation = "delete"
	OperationUpdate CrudOperation = "update"
	OperationRead   CrudOperation = "read"
)

// AllAllowedAccess is a slice of all CRUD operations.
var AllAllowedAccess = []CrudOperation{
	OperationCreate,
	OperationDelete,
	OperationUpdate,
	OperationRead,
}

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
