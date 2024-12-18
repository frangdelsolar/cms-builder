package builder_test

import (
	"fmt"
	"testing"

	"github.com/frangdelsolar/cms/builder"
	"github.com/stretchr/testify/assert"
)

func TestHasPermission(t *testing.T) {
	tests := []struct {
		name          string
		roles         []builder.Role
		action        builder.CrudOperation
		permissionMap builder.RolePermissionMap
		expected      bool
	}{
		{
			name:   "User has a role with the requested permission",
			roles:  []builder.Role{builder.AdminRole},
			action: builder.OperationCreate,
			permissionMap: builder.RolePermissionMap{
				builder.AdminRole: []builder.CrudOperation{builder.OperationCreate},
			},
			expected: true,
		},
		{
			name:   "User has multiple roles, one of which has the requested permission",
			roles:  []builder.Role{builder.VisitorRole, builder.AdminRole},
			action: builder.OperationCreate,
			permissionMap: builder.RolePermissionMap{
				builder.AdminRole: []builder.CrudOperation{builder.OperationCreate},
			},
			expected: true,
		},
		{
			name:   "User has multiple roles, none of which have the requested permission",
			roles:  []builder.Role{builder.VisitorRole, builder.SchedulerRole},
			action: builder.OperationCreate,
			permissionMap: builder.RolePermissionMap{
				builder.AdminRole: []builder.CrudOperation{builder.OperationCreate},
			},
			expected: false,
		},
		{
			name:   "User has a role that is not in the permission map",
			roles:  []builder.Role{builder.Role("unknown")},
			action: builder.OperationCreate,
			permissionMap: builder.RolePermissionMap{
				builder.AdminRole: []builder.CrudOperation{builder.OperationCreate},
			},
			expected: false,
		},
		{
			name:   "User has no roles",
			roles:  []builder.Role{},
			action: builder.OperationCreate,
			permissionMap: builder.RolePermissionMap{
				builder.AdminRole: []builder.CrudOperation{builder.OperationCreate},
			},
			expected: false,
		},
		{
			name:          "Permission map is empty",
			roles:         []builder.Role{builder.AdminRole},
			action:        builder.OperationCreate,
			permissionMap: builder.RolePermissionMap{},
			expected:      false,
		},
		{
			name:   "Action is not in the permission map",
			roles:  []builder.Role{builder.AdminRole},
			action: "unknown",
			permissionMap: builder.RolePermissionMap{
				builder.AdminRole: []builder.CrudOperation{builder.OperationCreate},
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
