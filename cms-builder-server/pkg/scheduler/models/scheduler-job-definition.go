package models

import (
	"time"

	authModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/models"
	schTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/scheduler/types"
)

type SchedulerJobDefinition struct {
	*authModels.SystemData
	Name          string                    `gorm:"not null,unique" json:"name"`
	FrequencyType schTypes.JobFrequencyType `json:"frequencyType"`
	AtTime        time.Time
	CronExpr      string `json:"cronExpr"`
	WithSeconds   bool   `json:"withSeconds"` // Cron expression with seconds
}
