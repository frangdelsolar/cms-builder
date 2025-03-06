package auth_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	authConstants "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/constants"
	authModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/models"
	authTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/types"
)

// TestGetIDString verifies that GetIDString returns the correct string representation of the user's ID.
func TestGetIDString(t *testing.T) {
	user := &authModels.User{ID: 123}
	assert.Equal(t, "123", user.StringID())
}

// TestGetRoles verifies that GetRoles correctly parses the comma-separated roles string.
func TestGetRoles(t *testing.T) {
	tests := []struct {
		name     string
		roles    string
		expected []authTypes.Role
	}{
		{
			name:     "Single role",
			roles:    "admin",
			expected: []authTypes.Role{authConstants.AdminRole},
		},
		{
			name:     "Multiple roles",
			roles:    "admin,visitor",
			expected: []authTypes.Role{authConstants.AdminRole, authConstants.VisitorRole},
		},
		{
			name:     "Empty roles",
			roles:    "",
			expected: []authTypes.Role{},
		},
		{
			name:     "Roles with spaces",
			roles:    "admin, visitor, scheduler",
			expected: []authTypes.Role{authConstants.AdminRole, authConstants.VisitorRole, authConstants.SchedulerRole},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &authModels.User{Roles: tt.roles}
			assert.Equal(t, tt.expected, user.GetRoles())
		})
	}
}

// TestSetRole verifies that SetRole correctly adds a role to the user's roles.
func TestSetRole(t *testing.T) {
	tests := []struct {
		name          string
		initialRoles  string
		roleToAdd     authTypes.Role
		expectedRoles string
		expectedError error
	}{
		{
			name:          "Add role to empty roles",
			initialRoles:  "",
			roleToAdd:     authConstants.AdminRole,
			expectedRoles: "admin",
			expectedError: nil,
		},
		{
			name:          "Add role to existing roles",
			initialRoles:  "admin",
			roleToAdd:     authConstants.VisitorRole,
			expectedRoles: "admin,visitor",
			expectedError: nil,
		},
		{
			name:          "Add duplicate role",
			initialRoles:  "admin,visitor",
			roleToAdd:     authConstants.AdminRole,
			expectedRoles: "admin,visitor",
			expectedError: authModels.ErrorRoleAlreadyAssigned,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &authModels.User{Roles: tt.initialRoles}
			err := user.SetRole(tt.roleToAdd)
			assert.Equal(t, tt.expectedError, err)
			assert.Equal(t, tt.expectedRoles, user.Roles)
		})
	}
}

// TestRemoveRole verifies that RemoveRole correctly removes a role from the user's roles.
func TestRemoveRole(t *testing.T) {
	tests := []struct {
		name          string
		initialRoles  string
		roleToRemove  authTypes.Role
		expectedRoles string
	}{
		{
			name:          "Remove existing role",
			initialRoles:  "admin,visitor",
			roleToRemove:  authConstants.AdminRole,
			expectedRoles: "visitor",
		},
		{
			name:          "Remove non-existent role",
			initialRoles:  "admin,visitor",
			roleToRemove:  authConstants.SchedulerRole,
			expectedRoles: "admin,visitor",
		},
		{
			name:          "Remove role from empty roles",
			initialRoles:  "",
			roleToRemove:  authConstants.AdminRole,
			expectedRoles: "",
		},
		{
			name:          "Remove last role",
			initialRoles:  "admin",
			roleToRemove:  authConstants.AdminRole,
			expectedRoles: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &authModels.User{Roles: tt.initialRoles}
			user.RemoveRole(tt.roleToRemove)
			assert.Equal(t, tt.expectedRoles, user.Roles)
		})
	}
}

func TestHasRole(t *testing.T) {
	tests := []struct {
		name     string
		user     authModels.User
		role     authTypes.Role
		expected bool
	}{
		{
			name:     "authModels.User has role",
			user:     authModels.User{Roles: "admin,visitor"},
			role:     authConstants.AdminRole,
			expected: true,
		},
		{
			name:     "authModels.User does not have role",
			user:     authModels.User{Roles: "user,visitor"},
			role:     authConstants.AdminRole,
			expected: false,
		},
		{
			name:     "authModels.User has no roles",
			user:     authModels.User{Roles: ""},
			role:     authConstants.AdminRole,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.user.HasRole(tt.role)
			assert.Equal(t, tt.expected, actual, "Expected %v but got %v", tt.expected, actual)
		})
	}
}
