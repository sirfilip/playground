package task_manager

import (
	"context"
	"fmt"
)

// Application service: operates outside of the domain
// works with domain objects and it is responsible for executing commands
// and persisting the aggregates through the repo

type Repository interface {
	BacklogByOwner(ctx context.Context, owner Owner) (Backlog, error)
	SaveBacklog(ctx context.Context, backlog Backlog) error
}

type TaskManager struct {
	repo Repository
}

func NewTaskManager(repo Repository) *TaskManager {
	return &TaskManager{repo: repo}
}

func (p *TaskManager) Schedule(ctx context.Context, cmd CreateTaskCmd) error {
	backlog, err := p.repo.BacklogByOwner(ctx, cmd.Owner)
	if err != nil {
		return fmt.Errorf("retreiving backlog: %w", err)
	}
	if err := backlog.Schedule(ctx, cmd); err != nil {
		return fmt.Errorf("backlog schedule: %w", err)
	}

	return p.repo.SaveBacklog(ctx, backlog)
}
