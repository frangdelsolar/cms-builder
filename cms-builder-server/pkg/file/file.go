package file

import (
	"net/http"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	manager "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/store"
)

func SetupFileResource(resourceManager *manager.ResourceManager, db *database.Database, st store.Store) *manager.ResourceConfig {

	skipUserBinding := false // DB Logs don't have a created_by field

	permissions := server.RolePermissionMap{
		models.AdminRole:   server.AllAllowedAccess,
		models.VisitorRole: server.AllAllowedAccess,
	}

	validators := manager.ValidatorsMap{
		"name": manager.ValidatorsList{manager.RequiredValidator},
		"path": manager.ValidatorsList{manager.RequiredValidator},
		"url":  manager.ValidatorsList{manager.RequiredValidator},
	}

	handlers := &manager.ApiHandlers{
		Create: CreateStoredFilesHandler(db, st),
		Delete: DeleteStoredFilesHandler(db, st),
		Update: UpdateStoredFilesHandler,
	}

	routes := []server.Route{
		{
			Path:         "/api/files/{id}/download",
			Handler:      DownloadStoredFileHandler(resourceManager, db, st),
			Name:         "files-download",
			RequiresAuth: true,
			Method:       http.MethodGet,
		},
	}

	config := &manager.ResourceConfig{
		Model:           models.DatabaseLog{},
		SkipUserBinding: skipUserBinding,
		Validators:      validators,
		Permissions:     permissions,
		Handlers:        handlers,
		Routes:          routes,
	}

	return config
}
