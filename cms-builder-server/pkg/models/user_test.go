package models_test

import (
	"testing"

	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	"github.com/stretchr/testify/assert"
)

// TestGetIDString verifies that GetIDString returns the correct string representation of the user's ID.
func TestGetIDString(t *testing.T) {
	user := &User{ID: 123}
	assert.Equal(t, "123", user.GetIDString())
}

// TestGetRoles verifies that GetRoles correctly parses the comma-separated roles string.
func TestGetRoles(t *testing.T) {
	tests := []struct {
		name     string
		roles    string
		expected []Role
	}{
		{
			name:     "Single role",
			roles:    "admin",
			expected: []Role{AdminRole},
		},
		{
			name:     "Multiple roles",
			roles:    "admin,visitor",
			expected: []Role{AdminRole, VisitorRole},
		},
		{
			name:     "Empty roles",
			roles:    "",
			expected: []Role{},
		},
		{
			name:     "Roles with spaces",
			roles:    "admin, visitor, scheduler",
			expected: []Role{AdminRole, VisitorRole, SchedulerRole},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{Roles: tt.roles}
			assert.Equal(t, tt.expected, user.GetRoles())
		})
	}
}

// TestSetRole verifies that SetRole correctly adds a role to the user's roles.
func TestSetRole(t *testing.T) {
	tests := []struct {
		name          string
		initialRoles  string
		roleToAdd     Role
		expectedRoles string
		expectedError error
	}{
		{
			name:          "Add role to empty roles",
			initialRoles:  "",
			roleToAdd:     AdminRole,
			expectedRoles: "admin",
			expectedError: nil,
		},
		{
			name:          "Add role to existing roles",
			initialRoles:  "admin",
			roleToAdd:     VisitorRole,
			expectedRoles: "admin,visitor",
			expectedError: nil,
		},
		{
			name:          "Add duplicate role",
			initialRoles:  "admin,visitor",
			roleToAdd:     AdminRole,
			expectedRoles: "admin,visitor",
			expectedError: ErrorRoleAlreadyAssigned,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{Roles: tt.initialRoles}
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
		roleToRemove  Role
		expectedRoles string
	}{
		{
			name:          "Remove existing role",
			initialRoles:  "admin,visitor",
			roleToRemove:  AdminRole,
			expectedRoles: "visitor",
		},
		{
			name:          "Remove non-existent role",
			initialRoles:  "admin,visitor",
			roleToRemove:  SchedulerRole,
			expectedRoles: "admin,visitor",
		},
		{
			name:          "Remove role from empty roles",
			initialRoles:  "",
			roleToRemove:  AdminRole,
			expectedRoles: "",
		},
		{
			name:          "Remove last role",
			initialRoles:  "admin",
			roleToRemove:  AdminRole,
			expectedRoles: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{Roles: tt.initialRoles}
			user.RemoveRole(tt.roleToRemove)
			assert.Equal(t, tt.expectedRoles, user.Roles)
		})
	}
}
