package task_manager

import (
	"time"

	"github.com/google/uuid"
)

type EventType string

const (
	TaskScheduled EventType = "TaskScheduled"
)

type EventID string

func NewEventID(id string) (EventID, error) {
	return EventID(id), nil
}

type EventSequence uint

type EventVersion uint

type AggregateID string

func (a AggregateID) Hash() (int64, error) {
	id, err := uuid.Parse(string(a))
	if err != nil {
		return 0, err
	}

	sec, _ := id.Time().UnixTime()
	return sec, nil
}

type Event struct {
	ID          EventID
	Type        EventType
	Sequence    EventSequence
	Version     EventVersion
	AggregateID AggregateID
	Payload     map[string]any
	Metadata    map[string]any
	Error       error
	Timestamp   time.Time
}
