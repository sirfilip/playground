package task_manager

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestTaskManagerScheduleBacklogExist(t *testing.T) {
	db := connect(t)
	t.Cleanup(func() {
		db.Close()
	})
	tm := NewTaskManager(&SqliteRepo{db: db})

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	taskId := TaskID(uuid.New().String())
	dueDate, err := NewDueDate(time.Now().Add(10 * time.Hour))
	if err != nil {
		t.Fatalf("failed to parse due date: %v", err)
	}

	if err := tm.Schedule(ctx, CreateTaskCmd{
		Owner:       Owner("filip"),
		Title:       TaskTitle("write some tests"),
		Description: TaskDescription("if you dont write tests you dont exist!"),
		DueDate:     dueDate,
		TaskId:      taskId,
	}); err != nil {
		t.Error(err)
	}

	var b Backlog
	if err := db.QueryRowContext(ctx, "select owner, version from backlogs").Scan(&b.Owner, &b.Version); err != nil {
		t.Error(err)
	}

	assertEquals(t, string(b.Owner), "filip")
	assertEquals(t, int(b.Version), 2)

	var task Task
	var dbDueDate string
	if err := db.QueryRowContext(
		ctx,
		`select
			version,
			task_id,
			title,
			owner,
			description,
			dueDate,
			status
		from tasks`,
	).Scan(
		&task.Version,
		&task.TaskID,
		&task.Title,
		&task.Owner,
		&task.Description,
		&dbDueDate,
		&task.Status,
	); err != nil {
		t.Error(err)
	}

	assertEquals(t, task.Version, TaskVersion(1))
	assertEquals(t, task.Owner, Owner("filip"))
	assertEquals(t, task.Title, TaskTitle("write some tests"))
	assertEquals(t, task.Description, TaskDescription("if you dont write tests you dont exist!"))

	// TODO add tests for the outbox
}

func TestTaskManagerScheduleBacklogDoesNotExist(t *testing.T) {
	t.Skip()
}
