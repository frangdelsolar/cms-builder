package auth

import (
	"net/http"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/clients"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	manager "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
)

func SetupUserResource(firebase *clients.FirebaseManager, db *database.Database, log *logger.Logger, getSystemUser func() *models.User) *manager.ResourceConfig {

	log.Info().Msg("Initializing User resource")

	skipUserBinding := true // Users don't have a created_by field

	permissions := server.RolePermissionMap{
		models.AdminRole:   server.AllAllowedAccess,
		models.VisitorRole: []server.CrudOperation{server.OperationRead},
	}

	validators := manager.ValidatorsMap{
		"Email": manager.ValidatorsList{manager.RequiredValidator, manager.EmailValidator},
		"Name":  manager.ValidatorsList{manager.RequiredValidator},
		"Roles": manager.ValidatorsList{manager.RequiredValidator},
	}

	// TODO: Should have its own handlers
	handlers := &manager.ApiHandlers{
		List:   nil, // Use default
		Detail: nil, // Use default
		Create: nil, // Use default
		Update: nil, // Use default
		Delete: nil, // Use default
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
