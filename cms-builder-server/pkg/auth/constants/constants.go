package auth

import (
	authTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/types"
)

// Predefined CRUD operations.
const (
	OperationCreate authTypes.CrudOperation = "create"
	OperationDelete authTypes.CrudOperation = "delete"
	OperationUpdate authTypes.CrudOperation = "update"
	OperationRead   authTypes.CrudOperation = "read"
)

// AllAllowedAccess is a slice of all CRUD operations.
var AllAllowedAccess = []authTypes.CrudOperation{
	OperationCreate,
	OperationDelete,
	OperationUpdate,
	OperationRead,
}

// Predefined user roles.
const (
	AdminRole     authTypes.Role = "admin"
	VisitorRole   authTypes.Role = "visitor"
	SchedulerRole authTypes.Role = "scheduler"
)

const GodTokenHeader = "X-God-Token"

const RolesParamKey = "roles"
