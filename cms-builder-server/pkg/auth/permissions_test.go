package auth_test

import (
	"fmt"
	"testing"

	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth"
	"github.com/stretchr/testify/assert"
)

func TestHasPermission(t *testing.T) {
	tests := []struct {
		name          string
		roles         []Role
		action        CrudOperation
		permissionMap RolePermissionMap
		expected      bool
	}{
		{
			name:   "User has a role with the requested permission",
			roles:  []Role{AdminRole},
			action: OperationCreate,
			permissionMap: RolePermissionMap{
				AdminRole: []CrudOperation{OperationCreate},
			},
			expected: true,
		},
		{
			name:   "User has multiple roles, one of which has the requested permission",
			roles:  []Role{VisitorRole, AdminRole},
			action: OperationCreate,
			permissionMap: RolePermissionMap{
				AdminRole: []CrudOperation{OperationCreate},
			},
			expected: true,
		},
		{
			name:   "User has multiple roles, none of which have the requested permission",
			roles:  []Role{VisitorRole, SchedulerRole},
			action: OperationCreate,
			permissionMap: RolePermissionMap{
				AdminRole: []CrudOperation{OperationCreate},
			},
			expected: false,
		},
		{
			name:   "User has a role that is not in the permission map",
			roles:  []Role{Role("unknown")},
			action: OperationCreate,
			permissionMap: RolePermissionMap{
				AdminRole: []CrudOperation{OperationCreate},
			},
			expected: false,
		},
		{
			name:   "User has no roles",
			roles:  []Role{},
			action: OperationCreate,
			permissionMap: RolePermissionMap{
				AdminRole: []CrudOperation{OperationCreate},
			},
			expected: false,
		},
		{
			name:          "Permission map is empty",
			roles:         []Role{AdminRole},
			action:        OperationCreate,
			permissionMap: RolePermissionMap{},
			expected:      false,
		},
		{
			name:   "Action is not in the permission map",
			roles:  []Role{AdminRole},
			action: "unknown",
			permissionMap: RolePermissionMap{
				AdminRole: []CrudOperation{OperationCreate},
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
