package file

import (
	"net/http"

	authConstants "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/constants"
	authTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/types"
	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"
	fileHandlers "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/file/handlers"
	fileModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/file/models"
	loggerTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger/types"
	rmPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"
	rmTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager/types"
	rmValidators "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager/validators"
	svrTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server/types"
	storeTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/store/types"
)

func SetupFileResource(resourceManager *rmPkg.ResourceManager, db *dbTypes.DatabaseConnection, st storeTypes.Store, log *loggerTypes.Logger, apiBaseUrl string) *rmTypes.ResourceConfig {

	log.Info().Msg("Initializing File resource")

	skipUserBinding := false // DB Logs don't have a created_by field

	permissions := authTypes.RolePermissionMap{
		authConstants.AdminRole:   authConstants.AllAllowedAccess,
		authConstants.VisitorRole: authConstants.AllAllowedAccess,
	}

	validators := rmTypes.ValidatorsMap{
		"Name": rmTypes.ValidatorsList{rmValidators.RequiredValidator},
	}

	handlers := &rmTypes.ApiHandlers{
		Create: fileHandlers.CreateStoredFilesHandler(db, st, apiBaseUrl),
		Delete: fileHandlers.DeleteStoredFilesHandler(db, st),
		Update: fileHandlers.UpdateStoredFilesHandler,
	}

	routes := []svrTypes.Route{
		{
			Path:         "/api/files/{id}/download",
			Handler:      fileHandlers.DownloadStoredFileHandler(resourceManager, db, st),
			Name:         "files-download",
			RequiresAuth: true,
			Methods:      []string{http.MethodGet, http.MethodHead},
		},
	}

	config := &rmTypes.ResourceConfig{
		Model:           fileModels.File{},
		SkipUserBinding: skipUserBinding,
		Validators:      validators,
		Permissions:     permissions,
		Handlers:        handlers,
		Routes:          routes,
	}

	return config
}
