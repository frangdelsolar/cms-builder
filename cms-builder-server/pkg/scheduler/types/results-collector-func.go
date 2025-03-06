package types

// ResultsCollectorFunc is a callback function to collect job results.
type ResultsCollectorFunc func(jobName string, results string)
