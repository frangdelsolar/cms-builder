package builder

const (
	OperationCreate CrudOperation = "create"
	OperationDelete CrudOperation = "delete"
	OperationUpdate CrudOperation = "update"
	OperationRead   CrudOperation = "read"

	AdminRole     Role = "admin"
	VisitorRole   Role = "visitor"
	SchedulerRole Role = "scheduler"
)

var AllAllowedAccess = []CrudOperation{
	OperationCreate,
	OperationDelete,
	OperationUpdate,
	OperationRead,
}

type Role string

func (r Role) S() string {
	return string(r)
}

type CrudOperation string

type RolePermissionMap map[Role][]CrudOperation

func (p RolePermissionMap) HasPermission(userRoles []Role, action CrudOperation) (
	isAllowed bool) {

	// Loop over the user's roles and their associated permissions
	for _, role := range userRoles {
		if _, ok := p[role]; ok {
			for _, allowedAction := range p[role] {
				if allowedAction == action {
					return true
				}
			}
		}
	}

	return false
}
