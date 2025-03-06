package models

import (
	authModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/models"
	schTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/scheduler/types"
)

type SchedulerTask struct {
	*authModels.SystemData
	JobDefinitionName string              `json:"jobDefinitionName"`
	Status            schTypes.TaskStatus `json:"status"`
	CronJobId         string              `json:"cronJobId"`
	Error             string              `json:"error"`
	Results           string              `json:"results"`
}
