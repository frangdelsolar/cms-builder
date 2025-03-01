package scheduler_test

import (
	"fmt"

	"github.com/go-co-op/gocron/v2"
)

// MockCron is a mock struct to simulate the behavior of the cron scheduler.
type MockCron struct {
	Task    gocron.Task
	Options []gocron.JobOption
}

func (m MockCron) Shutdown() error {
	fmt.Println("MockCron.Shutdown called")
	return nil
}

// NewJob is a mock method to simulate creating a new job.
func (m MockCron) NewJob(frequency gocron.JobDefinition, task gocron.Task, options ...gocron.JobOption) (gocron.Job, error) {
	fmt.Println("MockCron.NewJob called")
	fmt.Printf("Frequency: %v\n", frequency)
	fmt.Printf("Task: %v\n", task)
	for _, opt := range options {
		fmt.Printf("Option: %v\n", opt)
	}

	return nil, nil
}
