package task_manager

// Write model optimized for writing in db
// TODO add elastic search as projection that will be used
// as view model for filtering / searching tasks

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type SqliteRepo struct {
	db *sql.DB
}

func NewSqliteRepo(db *sql.DB) *SqliteRepo {
	return &SqliteRepo{
		db: db,
	}
}

func (r *SqliteRepo) BacklogByOwner(ctx context.Context, owner Owner) (Backlog, error) {
	b := Backlog{Owner: owner, Version: 1}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return b, err
	}
	defer tx.Rollback()
	if err := tx.QueryRowContext(
		ctx,
		"SELECT version FROM backlogs WHERE owner=? LIMIT 1",
		owner,
	).Scan(
		&b.Version,
	); err != nil {
		if err != sql.ErrNoRows {
			return b, fmt.Errorf("getting backlog version: %w", err)
		}

		if _, err := tx.ExecContext(ctx, "insert into backlogs (owner, version) values (?, ?)", b.Owner, 1); err != nil {
			return b, err
		}
	}

	var datetime string
	if err := tx.QueryRowContext(
		ctx,
		`select count(*), ifnull(min(dueDate), datetime('now'))
		from tasks where owner=? and status = ?`,
		owner,
		TaskStatusPending,
	).Scan(
		&b.PendingTasksCount, &datetime,
	); err != nil {
		return b, err
	}

	parsed, err := time.Parse("2006-01-02 15:04:05", datetime)
	if err != nil {
		return b, fmt.Errorf("parsing date: %v err: %w", datetime, err)
	}

	b.OldestTaskTime = parsed

	if err := tx.Commit(); err != nil {
		return b, err
	}

	return b, nil
}

func (r *SqliteRepo) SaveBacklog(ctx context.Context, backlog Backlog) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	res, err := tx.ExecContext(
		ctx,
		"UPDATE backlogs SET version = version + 1 where owner = ?",
		backlog.Owner,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	switch rowsAffected {
	case 0:
		return ErrOptimisticLockingFailure
	case 1:
		// pass
	default:
		// TODO inject logger into planner
		// more then one backlog per owner abort
		log.Printf("ERR: more then one backlog for owner: %s", backlog.Owner)
		return fmt.Errorf("%w: backlog save", ErrInvalidState)
	}

	for _, task := range backlog.NewTasks {
		if _, err := tx.ExecContext(ctx,
			`insert into tasks (
				version,
				task_id,
				title,
				owner,
				description,
				dueDate,
				createdAt,
				updatedAt,
				status
			) values (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			task.Version,
			task.TaskID,
			task.Title,
			task.Owner,
			task.Description,
			time.Time(task.DueDate),
			task.CreatedAt,
			task.UpdatedAt,
			task.Status,
		); err != nil {
			return err
		}
	}

	for _, event := range backlog.Outbox {
		payload, err := json.Marshal(event.Payload)
		if err != nil {
			return fmt.Errorf("marshal event payload: %w", err)
		}
		if _, err := tx.ExecContext(ctx,
			`insert into outbox (
				event_id,
				type,
				version,
				aggregate_id,
				payload,
				timestamp,
				status
			) values (?, ?, ?, ?, ?, ?, ?)`,
			event.ID,
			event.Type,
			event.Version,
			event.AggregateID,
			payload,
			event.Timestamp,
			OutboxStatusPending,
		); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("save backlog: %w", err)
	}

	return nil
}

type MemRepo struct {
	backlogs []Backlog
	tasks    []Task
	outbox   []Event
}
