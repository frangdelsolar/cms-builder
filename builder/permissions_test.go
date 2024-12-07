package builder_test

import (
	"fmt"
	"testing"

	"github.com/frangdelsolar/cms/builder"
	"github.com/stretchr/testify/assert"
)

var TestRole builder.Role = "testRole"

func TestGetQuery(t *testing.T) {
	tests := []struct {
		name               string
		roles              []builder.Role
		action             builder.PermissionAction
		params             builder.PermissionParams
		expectedQuery      string
		expectedFullAccess bool
		expectedError      error
	}{
		{
			name:               "Role not found",
			roles:              []builder.Role{"non-existent-role"},
			action:             builder.PermissionRead,
			params:             builder.PermissionParams{},
			expectedQuery:      "",
			expectedFullAccess: false,
			expectedError:      fmt.Errorf("no rules were found for action: read and user roles: [non-existent-role]"),
		},
		{
			name:               "Action not found",
			roles:              []builder.Role{"test-role"},
			action:             "non-existent-action",
			params:             builder.PermissionParams{},
			expectedQuery:      "",
			expectedFullAccess: false,
			expectedError:      fmt.Errorf("no rules were found for action: non-existent-action and user roles: [test-role]"),
		},
		{
			name:               "No filters found for action and role",
			roles:              []builder.Role{"test-role"},
			action:             builder.PermissionRead,
			params:             builder.PermissionParams{},
			expectedQuery:      "",
			expectedFullAccess: false,
			expectedError:      fmt.Errorf("no rules were found for action: read and user roles: [test-role]"),
		},
		{
			name:               "Multiple roles with permissions",
			roles:              []builder.Role{"test-role-1", "test-role-2"},
			action:             builder.PermissionRead,
			params:             builder.PermissionParams{"user_id": "1"},
			expectedQuery:      "user_id = '1'",
			expectedFullAccess: false,
			expectedError:      nil,
		},
		{
			name:               "Single role with multiple permissions",
			roles:              []builder.Role{"test-role"},
			action:             builder.PermissionRead,
			params:             builder.PermissionParams{"user_id": "1", "other_param": "2"},
			expectedQuery:      "user_id = '1' AND other_param = '2'",
			expectedFullAccess: false,
			expectedError:      nil,
		},
		{
			name:               "Empty params",
			roles:              []builder.Role{"test-role"},
			action:             builder.PermissionRead,
			params:             builder.PermissionParams{},
			expectedQuery:      "",
			expectedFullAccess: false,
			expectedError:      fmt.Errorf("no rules were found for action: read and user roles: [test-role]"),
		},
		{
			name:               "Params with empty values",
			roles:              []builder.Role{"test-role"},
			action:             builder.PermissionRead,
			params:             builder.PermissionParams{"user_id": "", "other_param": "2"},
			expectedQuery:      "other_param = '2'",
			expectedFullAccess: false,
			expectedError:      nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			permissionConfig := builder.RolePermissionMap{
				"test-role": builder.ActionToPermission{
					builder.PermissionRead: []builder.PermissionFilter{
						{
							FilteredFieldName: "user_id",
							ParameterKey:      "user_id",
						},
						{
							FilteredFieldName: "other_param",
							ParameterKey:      "other_param",
						},
					},
				},
				"test-role-1": builder.ActionToPermission{
					builder.PermissionRead: []builder.PermissionFilter{
						{
							FilteredFieldName: "user_id",
							ParameterKey:      "user_id",
						},
					},
				},
				"test-role-2": builder.ActionToPermission{
					builder.PermissionRead: []builder.PermissionFilter{
						{
							FilteredFieldName: "other_param",
							ParameterKey:      "other_param",
						},
					},
				},
			}

			fullAccess, query, err := permissionConfig.HasPermission(test.roles, test.action, test.params)

			if test.expectedError != nil {
				assert.EqualError(t, err, test.expectedError.Error())
			} else {
				assert.NoError(t, err, "GetQuery should not return an error")
			}

			assert.Equal(t, test.expectedQuery, query, "GetQuery should return the expected query")
			assert.Equal(t, test.expectedFullAccess, fullAccess, "GetQuery should return the expected hasPermission value")

		})
	}
}
