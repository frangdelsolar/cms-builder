package auth_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	authConstants "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/constants"
	authTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/types"
)

func TestRoleMapPermissionHasPermissions(t *testing.T) {
	tests := []struct {
		name          string
		roles         []authTypes.Role
		action        authTypes.CrudOperation
		permissionMap authTypes.RolePermissionMap
		expected      bool
	}{
		{
			name:   "User has a role with the requested permission",
			roles:  []authTypes.Role{authConstants.AdminRole},
			action: authConstants.OperationCreate,
			permissionMap: authTypes.RolePermissionMap{
				authConstants.AdminRole: []authTypes.CrudOperation{authConstants.OperationCreate},
			},
			expected: true,
		},
		{
			name:   "User has multiple roles, one of which has the requested permission",
			roles:  []authTypes.Role{authConstants.VisitorRole, authConstants.AdminRole},
			action: authConstants.OperationCreate,
			permissionMap: authTypes.RolePermissionMap{
				authConstants.AdminRole: []authTypes.CrudOperation{authConstants.OperationCreate},
			},
			expected: true,
		},
		{
			name:   "User has multiple roles, none of which have the requested permission",
			roles:  []authTypes.Role{authConstants.VisitorRole, authConstants.SchedulerRole},
			action: authConstants.OperationCreate,
			permissionMap: authTypes.RolePermissionMap{
				authConstants.AdminRole: []authTypes.CrudOperation{authConstants.OperationCreate},
			},
			expected: false,
		},
		{
			name:   "User has a role that is not in the permission map",
			roles:  []authTypes.Role{authTypes.Role("unknown")},
			action: authConstants.OperationCreate,
			permissionMap: authTypes.RolePermissionMap{
				authConstants.AdminRole: []authTypes.CrudOperation{authConstants.OperationCreate},
			},
			expected: false,
		},
		{
			name:   "User has no roles",
			roles:  []authTypes.Role{},
			action: authConstants.OperationCreate,
			permissionMap: authTypes.RolePermissionMap{
				authConstants.AdminRole: []authTypes.CrudOperation{authConstants.OperationCreate},
			},
			expected: false,
		},
		{
			name:          "Permission map is empty",
			roles:         []authTypes.Role{authConstants.AdminRole},
			action:        authConstants.OperationCreate,
			permissionMap: authTypes.RolePermissionMap{},
			expected:      false,
		},
		{
			name:   "Action is not in the permission map",
			roles:  []authTypes.Role{authConstants.AdminRole},
			action: "unknown",
			permissionMap: authTypes.RolePermissionMap{
				authConstants.AdminRole: []authTypes.CrudOperation{authConstants.OperationCreate},
			},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := test.permissionMap.HasPermission(test.roles, test.action)
			if actual != test.expected {
				assert.Equal(t, test.expected, actual, fmt.Sprintf("expected: %v, actual: %v", test.expected, actual))
			}
		})
	}
}
