package constants

import (
	schTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/scheduler/types"
)

const (
	TaskStatusRunning schTypes.TaskStatus = "running" // Task is currently running.
	TaskStatusFailed  schTypes.TaskStatus = "failed"  // Task failed during execution.
	TaskStatusDone    schTypes.TaskStatus = "done"    // Task completed successfully.
)

const (
	JobFrequencyTypeImmediate schTypes.JobFrequencyType = "immediate" // Job runs immediately.
	JobFrequencyTypeScheduled schTypes.JobFrequencyType = "scheduled" // Job runs at a specific time.
	JobFrequencyTypeCron      schTypes.JobFrequencyType = "cron"      // Job runs based on a cron expression.
)
