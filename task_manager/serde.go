package task_manager

import (
	"encoding/json"
	"fmt"
	"time"
)

type SerdeDBEvent struct {
	ID          string
	Type        string
	Sequence    int
	Version     int
	AggregateID string
	Payload     string
	Metadata    string
	Timestamp   string
}

func (serde SerdeDBEvent) Event() (Event, error) {
	event := Event{}

	if err := json.Unmarshal([]byte(serde.Payload), &event.Payload); err != nil {
		return event, fmt.Errorf("serde unmarshal payload: %w", err)
	}

	if err := json.Unmarshal([]byte(serde.Metadata), &event.Metadata); err != nil {
		return event, fmt.Errorf("serde unmarshal metadata: %w", err)
	}

	return event, nil
}

func NewSerdeDBEvent(event Event) (SerdeDBEvent, error) {
	serde := SerdeDBEvent{}
	payload, err := json.Marshal(event.Payload)
	if err != nil {
		return serde, fmt.Errorf("serde event to db: json marshal payload: %w", err)
	}
	serde.Payload = string(payload)
	metadata, err := json.Marshal(event.Metadata)
	if err != nil {
		return serde, fmt.Errorf("serde event to db: json marshal metadata: %w", err)
	}
	serde.Metadata = string(metadata)
	serde.ID = string(event.ID)
	serde.Type = string(event.Type)
	serde.Sequence = int(event.Sequence)
	serde.Version = int(event.Version)
	serde.AggregateID = string(event.AggregateID)
	serde.Timestamp = event.Timestamp.Format(time.RFC3339)

	return serde, nil
}

type SerdeJSONEvent struct {
	ID          string         `json:"id"`
	Type        string         `json:"type"`
	Sequence    int            `json:"sequence"`
	Version     int            `json:"version"`
	AggregateID string         `json:"aggregate_id"`
	Payload     map[string]any `json:"payload"`
	Metadata    map[string]any `json:"metadata"`
	Timestamp   string         `json:"timestamp"`
}

func (serde SerdeJSONEvent) Event() (Event, error) {
	event := Event{}

	return event, nil
}

func NewSerdeJSONEvent(event Event) (SerdeJSONEvent, error) {
	serde := SerdeJSONEvent{}
	serde.Payload = event.Payload
	serde.Metadata = event.Metadata
	serde.ID = string(event.ID)
	serde.Type = string(event.Type)
	serde.Sequence = int(event.Sequence)
	serde.Version = int(event.Version)
	serde.AggregateID = string(event.AggregateID)
	serde.Timestamp = event.Timestamp.Format(time.RFC3339)

	return serde, nil
}
