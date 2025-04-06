package task_manager

import "time"

type TaskStatus int

const (
	_ TaskStatus = iota
	TaskStatusPending
	TaskStatusCompleted
	TaskStatusRejected
)

const (
	_ TaskStatus = iota
	Pending
	Completed
	Rejected
)

type Task struct {
	Version     TaskVersion
	TaskID      TaskID
	Owner       Owner
	Title       TaskTitle
	Description TaskDescription
	DueDate     DueDate
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Status      TaskStatus
}
