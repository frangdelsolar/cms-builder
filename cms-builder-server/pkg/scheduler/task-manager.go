package scheduler

import "sync"

type TaskManager struct {
	sync.RWMutex
	Tasks map[string]string // jobName -> results
}

func (tm *TaskManager) Set(jobName, results string) {
	tm.Lock()
	defer tm.Unlock()
	tm.Tasks[jobName] = results
}

func (tm *TaskManager) Get(jobName string) (string, bool) {
	tm.RLock()
	defer tm.RUnlock()
	results, ok := tm.Tasks[jobName]
	return results, ok
}

func (tm *TaskManager) Delete(jobName string) {
	tm.Lock()
	defer tm.Unlock()
	delete(tm.Tasks, jobName)
}
