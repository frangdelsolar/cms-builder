package file

import (
	"net/http"

	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"
	loggerTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger/types"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	manager "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/store"
)

func SetupFileResource(resourceManager *manager.ResourceManager, db *dbTypes.DatabaseConnection, st store.Store, log *loggerTypes.Logger, apiBaseUrl string) *manager.ResourceConfig {

	log.Info().Msg("Initializing File resource")

	skipUserBinding := false // DB Logs don't have a created_by field

	permissions := server.RolePermissionMap{
		models.AdminRole:   server.AllAllowedAccess,
		models.VisitorRole: server.AllAllowedAccess,
	}

	validators := manager.ValidatorsMap{
		"Name": manager.ValidatorsList{manager.RequiredValidator},
	}

	handlers := &manager.ApiHandlers{
		Create: CreateStoredFilesHandler(db, st, apiBaseUrl),
		Delete: DeleteStoredFilesHandler(db, st),
		Update: UpdateStoredFilesHandler,
	}

	routes := []server.Route{
		{
			Path:         "/api/files/{id}/download",
			Handler:      DownloadStoredFileHandler(resourceManager, db, st),
			Name:         "files-download",
			RequiresAuth: true,
			Methods:      []string{http.MethodGet, http.MethodHead},
		},
	}

	config := &manager.ResourceConfig{
		Model:           models.File{},
		SkipUserBinding: skipUserBinding,
		Validators:      validators,
		Permissions:     permissions,
		Handlers:        handlers,
		Routes:          routes,
	}

	return config
}
