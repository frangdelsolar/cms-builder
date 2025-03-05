package scheduler

import (
	"time"

	authModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/models"
)

type TaskStatus string

type SchedulerTask struct {
	*authModels.SystemData
	JobDefinitionName string     `json:"jobDefinitionName"`
	Status            TaskStatus `json:"status"`
	CronJobId         string     `json:"cronJobId"`
	Error             string     `json:"error"`
	Results           string     `json:"results"`
}

type JobFrequencyType string

type SchedulerJobDefinition struct {
	*authModels.SystemData
	Name          string           `gorm:"not null,unique" json:"name"`
	FrequencyType JobFrequencyType `json:"frequencyType"`
	AtTime        time.Time
	CronExpr      string `json:"cronExpr"`
	WithSeconds   bool   `json:"withSeconds"` // Cron expression with seconds
}
