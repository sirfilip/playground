package task_manager

type OutboxStatus uint

const (
	_ OutboxStatus = iota
	OutboxStatusPending
	OutboxStatusDelivered
	OutboxStatusFailed
)
