package middlewarestest_test

import (
	"testing"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
	"github.com/stretchr/testify/assert"
)

func TestFormatRoles(t *testing.T) {
	tests := []struct {
		name     string
		roles    []models.Role
		expected string
	}{
		{
			name:     "single role",
			roles:    []models.Role{models.AdminRole},
			expected: "admin",
		},
		{
			name:     "multiple roles",
			roles:    []models.Role{models.AdminRole, models.VisitorRole, models.SchedulerRole},
			expected: "admin,visitor,scheduler",
		},
		{
			name:     "empty roles",
			roles:    []models.Role{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatRoles(tt.roles)
			assert.Equal(t, tt.expected, result)
		})
	}
}
