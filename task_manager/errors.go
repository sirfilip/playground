package task_manager

type TaskManagerError string

func (p TaskManagerError) Error() string {
	return string(p)
}

var _ error = TaskManagerError("")

const (
	ErrValue                    TaskManagerError = "value error"
	ErrNotFound                 TaskManagerError = "not found"
	ErrScheduleFull             TaskManagerError = "schedule is full"
	ErrScheduleStale            TaskManagerError = "schedule stale error"
	ErrOptimisticLockingFailure TaskManagerError = "optimistic locking failure. reload and try again"
	ErrInvalidState             TaskManagerError = "invalid state"
)
