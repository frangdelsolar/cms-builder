package builder

import "fmt"

const (
	PermissionCreate PermissionAction = "create"
	PermissionDelete PermissionAction = "delete"
	PermissionUpdate PermissionAction = "update"
	PermissionRead   PermissionAction = "read"

	AdminRole         Role = "admin"
	VisitorRole       Role = "visitor"
	SchedulerRole     Role = "scheduler"
	AuthenticatorRole Role = "authenticator"

	createdByIdField = "created_by_id"
)

var AllAllowedAccess = ActionToPermission{
	PermissionCreate: []PermissionFilter{
		{
			FullAccess: true,
		},
	},
	PermissionRead: []PermissionFilter{
		{
			FullAccess: true,
		},
	},
	PermissionUpdate: []PermissionFilter{
		{
			FullAccess: true,
		},
	},
	PermissionDelete: []PermissionFilter{
		{
			FullAccess: true,
		},
	},
}

var OwnerAccess = ActionToPermission{
	PermissionCreate: []PermissionFilter{
		{
			FullAccess: true,
		},
	},
	PermissionRead: []PermissionFilter{
		{
			FilteredFieldName: createdByIdField,
			ParameterKey:      requestedByParamKey,
		},
	},
	PermissionUpdate: []PermissionFilter{
		{
			FilteredFieldName: createdByIdField,
			ParameterKey:      requestedByParamKey,
		},
	},
	PermissionDelete: []PermissionFilter{
		{
			FilteredFieldName: createdByIdField,
			ParameterKey:      requestedByParamKey,
		},
	},
}

type Role string
type PermissionAction string

type PermissionFilter struct {
	FilteredFieldName FieldName       `json:"filteredFieldName"`
	ParameterKey      RequestParamKey `json:"mapedFieldName"`
	FullAccess        bool            `json:"fullAccess"`
}
type ActionToPermission map[PermissionAction][]PermissionFilter
type RolePermissionMap map[Role]ActionToPermission

func (p RolePermissionMap) HasPermission(userRoles []Role, action PermissionAction, params RequestParameters) (
	fullAccess bool, query string, err error) {

	query = ""
	err = nil
	fullAccess = false

	// Loop over the user's roles and their associated permissions
	for _, role := range userRoles {

		// Check if the role has permissions for the specified action
		actionMap, ok := p[role]
		if !ok {
			continue
		}

		// Check if there are any filters associated with the permission
		filters, ok := actionMap[action]
		if ok {

			// If no filters are associated with the permission, continue to the next role
			if len(filters) == 0 {
				continue
			}

			// Review the filters and add them to the query
			for _, filter := range filters {

				// If the filter is full access, set the fullAccess flag to true
				// and return
				if filter.FullAccess {
					return true, "", nil
				}

				if params[filter.ParameterKey] != "" {
					if query != "" {
						query += " AND "
					}
					query += filter.FilteredFieldName.S() + " = '" + params[filter.ParameterKey] + "'"
				}
			}

			if query != "" {
				return false, query, nil
			}
		}

	}

	return false, "", fmt.Errorf("no rules were found for action: %s and user roles: %v", action, userRoles)
}
