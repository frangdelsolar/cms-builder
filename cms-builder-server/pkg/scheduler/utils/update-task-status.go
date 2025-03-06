package utils

import (
	"context"

	authModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/models"
	dbQueries "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/queries"
	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"
	loggerTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger/types"
	schTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/scheduler/types"
	utilsPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/utils"
)

func UpdateTaskStatus(log *loggerTypes.Logger, db *dbTypes.DatabaseConnection, schedulerUser *authModels.User, cronJobId string, status schTypes.TaskStatus, errMsg string, requestId string, results string) error {
	task := GetSchedulerTask(log, db, cronJobId)
	task.SystemData.UpdatedByID = schedulerUser.ID
	task.Status = status
	task.Error = errMsg
	task.Results = results

	previousState := GetSchedulerTask(log, db, cronJobId)
	differences := utilsPkg.CompareInterfaces(previousState, task)
	return dbQueries.Update(context.Background(), log, db, task, schedulerUser, differences, requestId)
}
