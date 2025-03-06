package types

type JobRegistryTaskDefinition struct {
	Function   SchedulerTaskFunc // The task function to execute.
	Parameters []any             // Parameters for the task function.
}
