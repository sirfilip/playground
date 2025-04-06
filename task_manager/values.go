package task_manager

import (
	"fmt"
	"time"
)

type TaskVersion uint64

type TaskID string

func NewTaskID(taskID string) (TaskID, error) {
	return TaskID(taskID), nil
}

type Owner string

func NewOwner(owner string) (Owner, error) {
	return Owner(owner), nil
}

type TaskTitle string

func NewTaskTitle(taskTitle string) (TaskTitle, error) {
	if len(taskTitle) > 100 {
		return TaskTitle(""), fmt.Errorf("%w: task must be under 100 characters long", ErrValue)
	}

	return TaskTitle(taskTitle), nil
}

type TaskDescription string

func NewTaskDescription(desc string) (TaskDescription, error) {
	return TaskDescription(desc), nil
}

type DueDate time.Time

var NullDueDate = DueDate(time.UnixMilli(0))

func NewDueDate(dueDate time.Time) (DueDate, error) {
	if dueDate.Before(time.Now()) {
		return NullDueDate, fmt.Errorf("due date must be in the future")
	}

	return DueDate(dueDate), nil
}
