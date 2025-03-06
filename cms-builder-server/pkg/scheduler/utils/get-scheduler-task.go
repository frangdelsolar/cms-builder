package utils

import (
	"context"

	dbQueries "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/queries"
	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"
	loggerTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger/types"
	schModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/scheduler/models"
)

func GetSchedulerTask(log *loggerTypes.Logger, db *dbTypes.DatabaseConnection, cronJobId string) *schModels.SchedulerTask {
	var task schModels.SchedulerTask

	filters := map[string]interface{}{
		"cron_job_id": cronJobId,
	}

	err := dbQueries.FindOne(context.Background(), log, db, &task, filters)
	if err != nil {
		return nil
	}

	return &task
}
