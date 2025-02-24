package models

import "time"

type JobFrequencyType string

type SchedulerJobDefinition struct {
	*SystemData
	Name          string           `gorm:"not null,unique" json:"name"`
	FrequencyType JobFrequencyType `json:"frequencyType"`
	AtTime        time.Time
	CronExpr      string `json:"cronExpr"`
	WithSeconds   bool   `json:"withSeconds"` // Cron expression with seconds
}
