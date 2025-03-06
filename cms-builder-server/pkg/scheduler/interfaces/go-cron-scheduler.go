package interfaces

import "github.com/go-co-op/gocron/v2"

// GoCronScheduler is an interface for interacting with the gocron scheduler.
type GoCronScheduler interface {
	// NewJob creates a new job with the given frequency, task, and event listeners.
	NewJob(frequency gocron.JobDefinition, task gocron.Task, eventListeners ...gocron.JobOption) (gocron.Job, error)
	// Shutdown stops the scheduler.
	Shutdown() error
}
