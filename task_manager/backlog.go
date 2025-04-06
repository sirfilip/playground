package task_manager

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const MAX_PENDING_TASKS = 10
const MAX_TICKET_DEADLINE_HOURS = 24 * 6 * 30

// TODO use values for pending tasks count to guard the invariant
type Backlog struct {
	Owner             Owner
	PendingTasksCount uint
	OldestTaskTime    time.Time
	Version           uint
	NewTasks          []Task
	ClosedTasks       []TaskID
	Outbox            []Event
}

func (b *Backlog) Schedule(ctx context.Context, cmd CreateTaskCmd) error {
	task := Task{
		Version:     1,
		TaskID:      cmd.TaskId,
		Owner:       cmd.Owner,
		Title:       cmd.Title,
		Description: cmd.Description,
		DueDate:     cmd.DueDate,
	}

	now := time.Now()

	if now.Sub(b.OldestTaskTime).Abs().Hours() > MAX_TICKET_DEADLINE_HOURS {
		return ErrScheduleStale
	}

	if b.PendingTasksCount > MAX_PENDING_TASKS {
		return ErrScheduleFull
	}

	task.CreatedAt = now
	task.UpdatedAt = now
	b.PendingTasksCount += 1
	b.NewTasks = append(b.NewTasks, task)

	id, err := uuid.NewRandom()
	if err != nil {
		return fmt.Errorf("scheudle: uuid gen: %w", err)
	}

	eventID, err := NewEventID(id.String())
	if err != nil {
		return fmt.Errorf("eventId from uuid: %w", err)
	}

	b.Outbox = append(b.Outbox, Event{
		ID:          eventID,
		Type:        TaskScheduled,
		AggregateID: AggregateID(task.TaskID),
		Version:     1,
		Payload: map[string]any{
			"task": task,
		},
		Timestamp: now,
	})
	return nil
}
