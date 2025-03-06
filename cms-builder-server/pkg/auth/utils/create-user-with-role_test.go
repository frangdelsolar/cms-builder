package auth_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	authConstants "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/constants"
	authTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/types"
	authUtils "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/utils"
)

func TestFormatRoles(t *testing.T) {
	tests := []struct {
		name     string
		roles    []authTypes.Role
		expected string
	}{
		{
			name:     "single role",
			roles:    []authTypes.Role{authConstants.AdminRole},
			expected: "admin",
		},
		{
			name:     "multiple roles",
			roles:    []authTypes.Role{authConstants.AdminRole, authConstants.VisitorRole, authConstants.SchedulerRole},
			expected: "admin,visitor,scheduler",
		},
		{
			name:     "empty roles",
			roles:    []authTypes.Role{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := authUtils.FormatRoles(tt.roles)
			assert.Equal(t, tt.expected, result)
		})
	}
}
