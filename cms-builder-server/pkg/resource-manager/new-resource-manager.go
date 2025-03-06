package resourcemanager

import (
	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"
	loggerTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger/types"
	rmTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager/types"
)

func NewResourceManager(db *dbTypes.DatabaseConnection, log *loggerTypes.Logger) *ResourceManager {
	return &ResourceManager{
		Resources: make(map[string]*rmTypes.Resource),
		DB:        db,
		Logger:    log,
	}
}
