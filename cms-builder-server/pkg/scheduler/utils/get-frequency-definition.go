package utils

import (
	"fmt"

	schConstants "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/scheduler/constants"
	schModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/scheduler/models"
	"github.com/go-co-op/gocron/v2"
)

func GetFrequencyDefinition(jobDefinition *schModels.SchedulerJobDefinition) (gocron.JobDefinition, error) {

	switch jobDefinition.FrequencyType {
	case schConstants.JobFrequencyTypeImmediate:
		return gocron.OneTimeJob(
			gocron.OneTimeJobStartImmediately(),
		), nil

	case schConstants.JobFrequencyTypeScheduled:
		return gocron.OneTimeJob(
			gocron.OneTimeJobStartDateTimes(jobDefinition.AtTime),
		), nil

	case schConstants.JobFrequencyTypeCron:
		if jobDefinition.CronExpr == "" {
			return nil, fmt.Errorf("cron expression is required")
		}

		return gocron.CronJob(
			jobDefinition.CronExpr,
			jobDefinition.WithSeconds,
		), nil

	}
	return nil, fmt.Errorf("unknown frequency type: %s", jobDefinition.FrequencyType)
}
