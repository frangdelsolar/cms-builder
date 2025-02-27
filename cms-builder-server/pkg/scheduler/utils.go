package scheduler

import (
	"fmt"

	"github.com/go-co-op/gocron/v2"
)

func getFrequencyDefinition(jobDefinition *SchedulerJobDefinition) (gocron.JobDefinition, error) {

	switch jobDefinition.FrequencyType {
	case JobFrequencyTypeImmediate:
		return gocron.OneTimeJob(
			gocron.OneTimeJobStartImmediately(),
		), nil

	case JobFrequencyTypeScheduled:
		return gocron.OneTimeJob(
			gocron.OneTimeJobStartDateTimes(jobDefinition.AtTime),
		), nil

	case JobFrequencyTypeCron:
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
