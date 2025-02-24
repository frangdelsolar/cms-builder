package models

type TaskStatus string

type SchedulerTask struct {
	*SystemData
	JobDefinitionName string     `json:"jobDefinitionName"`
	Status            TaskStatus `json:"status"`
	CronJobId         string     `json:"cronJobId"`
	Error             string     `json:"error"`
}
