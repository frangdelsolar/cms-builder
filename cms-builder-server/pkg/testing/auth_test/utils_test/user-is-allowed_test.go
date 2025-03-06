package auth_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	authConstants "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/constants"
	authTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/types"
	authUtils "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/utils"
	loggerPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger"
)

// TestUserIsAllowed tests the UserIsAllowed function.
func TestUserIsAllowed(t *testing.T) {

	const EditorRole authTypes.Role = "editor"
	// Define test cases
	tests := []struct {
		name           string
		appPermissions authTypes.RolePermissionMap
		userRoles      []authTypes.Role
		action         authTypes.CrudOperation
		expectedResult bool
	}{
		{
			name: "user has permission for the action",
			appPermissions: authTypes.RolePermissionMap{
				authConstants.AdminRole: {authConstants.OperationRead, authConstants.OperationCreate},
				EditorRole:              {authConstants.OperationRead},
			},
			userRoles:      []authTypes.Role{authConstants.AdminRole},
			action:         authConstants.OperationRead,
			expectedResult: true,
		},
		{
			name: "user does not have permission for the action",
			appPermissions: authTypes.RolePermissionMap{
				authConstants.AdminRole: {authConstants.OperationCreate},
				EditorRole:              {authConstants.OperationRead},
			},
			userRoles:      []authTypes.Role{EditorRole},
			action:         authConstants.OperationCreate,
			expectedResult: false,
		},
		{
			name: "user has no roles",
			appPermissions: authTypes.RolePermissionMap{
				authConstants.AdminRole: {authConstants.OperationRead, authConstants.OperationCreate},
				EditorRole:              {authConstants.OperationRead},
			},
			userRoles:      []authTypes.Role{},
			action:         authConstants.OperationRead,
			expectedResult: false,
		},
		{
			name: "role has no permissions",
			appPermissions: authTypes.RolePermissionMap{
				authConstants.AdminRole: {},
				EditorRole:              {authConstants.OperationRead},
			},
			userRoles:      []authTypes.Role{authConstants.AdminRole},
			action:         authConstants.OperationRead,
			expectedResult: false,
		},
		{
			name: "role does not exist in permissions",
			appPermissions: authTypes.RolePermissionMap{
				EditorRole: {authConstants.OperationRead},
			},
			userRoles:      []authTypes.Role{authConstants.AdminRole},
			action:         authConstants.OperationRead,
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the function
			result := authUtils.UserIsAllowed(tt.appPermissions, tt.userRoles, tt.action, "test-app-name", loggerPkg.Default)

			// Verify the result
			assert.Equal(t, tt.expectedResult, result, "Unexpected result for test case: %s", tt.name)
		})
	}
}
