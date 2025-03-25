package auth

import (
	"net/http"

	authConstants "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/constants"
	authHandlers "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/handlers"
	authModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/models"
	authTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/types"
	cliPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/clients"
	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"
	loggerTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger/types"
	rmTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager/types"
	rmValidators "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager/validators"
	svrTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server/types"
)

func SetupUserResource(firebase *cliPkg.FirebaseManager, db *dbTypes.DatabaseConnection, log *loggerTypes.Logger, getSystemUser func() *authModels.User) *rmTypes.ResourceConfig {

	log.Info().Msg("Initializing User resource")

	skipUserBinding := true // Users don't have a created_by field

	permissions := authTypes.RolePermissionMap{
		authConstants.AdminRole:   authConstants.AllAllowedAccess,
		authConstants.VisitorRole: []authTypes.CrudOperation{authConstants.OperationRead},
	}

	validators := rmTypes.ValidatorsMap{
		"Email":     rmTypes.ValidatorsList{rmValidators.RequiredValidator, rmValidators.EmailValidator},
		"FirstName": rmTypes.ValidatorsList{rmValidators.RequiredValidator},
	}

	handlers := &rmTypes.ApiHandlers{
		List:   authHandlers.UserListHandler,
		Detail: authHandlers.UserDetailHandler,
		Create: authHandlers.UserCreateHandler,
		Update: authHandlers.UserUpdateHandler,
		Delete: authHandlers.UserDeleteHandler,
	}

	routes := []svrTypes.Route{
		{
			Path:         "/auth/register",
			Handler:      authHandlers.RegisterVisitorController(firebase, db, getSystemUser),
			Name:         "register",
			RequiresAuth: false,
			Methods:      []string{http.MethodPost},
		},
	}

	config := &rmTypes.ResourceConfig{
		Model:           authModels.User{},
		SkipUserBinding: skipUserBinding,
		Validators:      validators,
		Permissions:     permissions,
		Handlers:        handlers,
		Routes:          routes,
	}

	return config
}
