package types

// SchedulerTaskFunc is a function to execute a job.
type SchedulerTaskFunc func(jobParameters ...any) (string, error) // results, error
