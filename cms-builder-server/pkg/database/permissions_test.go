package database_test

import (
	"fmt"
	"testing"

	pkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	"github.com/stretchr/testify/assert"
)

func TestHasPermission(t *testing.T) {
	tests := []struct {
		name          string
		roles         []pkg.Role
		action        pkg.CrudOperation
		permissionMap pkg.RolePermissionMap
		expected      bool
	}{
		{
			name:   "User has a role with the requested permission",
			roles:  []pkg.Role{pkg.AdminRole},
			action: pkg.OperationCreate,
			permissionMap: pkg.RolePermissionMap{
				pkg.AdminRole: []pkg.CrudOperation{pkg.OperationCreate},
			},
			expected: true,
		},
		{
			name:   "User has multiple roles, one of which has the requested permission",
			roles:  []pkg.Role{pkg.VisitorRole, pkg.AdminRole},
			action: pkg.OperationCreate,
			permissionMap: pkg.RolePermissionMap{
				pkg.AdminRole: []pkg.CrudOperation{pkg.OperationCreate},
			},
			expected: true,
		},
		{
			name:   "User has multiple roles, none of which have the requested permission",
			roles:  []pkg.Role{pkg.VisitorRole, pkg.SchedulerRole},
			action: pkg.OperationCreate,
			permissionMap: pkg.RolePermissionMap{
				pkg.AdminRole: []pkg.CrudOperation{pkg.OperationCreate},
			},
			expected: false,
		},
		{
			name:   "User has a role that is not in the permission map",
			roles:  []pkg.Role{pkg.Role("unknown")},
			action: pkg.OperationCreate,
			permissionMap: pkg.RolePermissionMap{
				pkg.AdminRole: []pkg.CrudOperation{pkg.OperationCreate},
			},
			expected: false,
		},
		{
			name:   "User has no roles",
			roles:  []pkg.Role{},
			action: pkg.OperationCreate,
			permissionMap: pkg.RolePermissionMap{
				pkg.AdminRole: []pkg.CrudOperation{pkg.OperationCreate},
			},
			expected: false,
		},
		{
			name:          "Permission map is empty",
			roles:         []pkg.Role{pkg.AdminRole},
			action:        pkg.OperationCreate,
			permissionMap: pkg.RolePermissionMap{},
			expected:      false,
		},
		{
			name:   "Action is not in the permission map",
			roles:  []pkg.Role{pkg.AdminRole},
			action: "unknown",
			permissionMap: pkg.RolePermissionMap{
				pkg.AdminRole: []pkg.CrudOperation{pkg.OperationCreate},
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
