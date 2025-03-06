package types

import (
	authTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/types"
	svrTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server/types"
)

type ResourceConfig struct {
	Model           interface{}
	Handlers        *ApiHandlers
	SkipUserBinding bool
	Validators      ValidatorsMap
	Permissions     authTypes.RolePermissionMap
	Routes          []svrTypes.Route
}
