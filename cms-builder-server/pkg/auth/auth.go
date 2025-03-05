package auth

import (
	"net/http"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/clients"
	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"
	loggerTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger/types"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	manager "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
)

func SetupUserResource(firebase *clients.FirebaseManager, db *dbTypes.DatabaseConnection, log *loggerTypes.Logger, getSystemUser func() *models.User) *manager.ResourceConfig {

	log.Info().Msg("Initializing User resource")

	skipUserBinding := true // Users don't have a created_by field

	permissions := server.RolePermissionMap{
		models.AdminRole:   server.AllAllowedAccess,
		models.VisitorRole: []server.CrudOperation{server.OperationRead},
	}

	validators := manager.ValidatorsMap{
		"Email": manager.ValidatorsList{manager.RequiredValidator, manager.EmailValidator},
		"Name":  manager.ValidatorsList{manager.RequiredValidator},
	}

	handlers := &manager.ApiHandlers{
		List:   UserListHandler,
		Detail: UserDetailHandler,
		Create: UserCreateHandler,
		Update: UserUpdateHandler,
		Delete: UserDeleteHandler,
	}

	routes := []server.Route{
		{
			Path:         "/auth/register",
			Handler:      RegisterVisitorController(firebase, db, getSystemUser),
			Name:         "register",
			RequiresAuth: false,
			Methods:      []string{http.MethodPost},
		},
	}

	config := &manager.ResourceConfig{
		Model:           models.User{},
		SkipUserBinding: skipUserBinding,
		Validators:      validators,
		Permissions:     permissions,
		Handlers:        handlers,
		Routes:          routes,
	}

	return config
}
