package task_manager

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strings"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

const MaxAttempts = 3

type Publisher struct {
	workers int
	db      *sql.DB
	broker  *amqp.Connection
	logger  *slog.Logger
}

// TODO add failed reason in the outbox
func (p *Publisher) Publish(ctx context.Context) error {
	ch, err := p.broker.Channel()
	if err != nil {
		return fmt.Errorf("publish channel: %w", err)
	}
	defer ch.Close()

	if err := ch.ExchangeDeclare(
		"task_management_topic", // name
		"topic",                 // type
		true,                    // durable
		false,                   // auto-deleted
		false,                   // internal
		false,                   // no-wait
		nil,                     // arguments
	); err != nil {
		return fmt.Errorf("publish declare exchange: %w", err)
	}

	lockStmt, err := p.db.PrepareContext(ctx, `
		UPDATE outbox SET 
			reserved_by = ?
			reserved_at = current_timestamp
			attempt = attempt + 1
		WHERE
			event_id in (
				SELECT event_id FROM outbox
				WHERE
					attempt < 5 AND (
					(status = 1 AND reserved_by = NULL) 
					OR
					(status = 1 AND reserved_by = ?)
					OR
					(status = 1 AND reserved_by <> ? AND (strftime('%s', datetime('now')) - strftime('%s', reserved_at)) > 300))
				LIMIT 100
			)
	`)
	if err != nil {
		return fmt.Errorf("publish create lock stmt: %w", err)
	}
	defer lockStmt.Close()

	lock := func(ctx context.Context, ident string) error {
		if _, err := lockStmt.ExecContext(ctx, ident); err != nil {
			return fmt.Errorf("exec lock: %w", err)
		}
		return nil
	}

	lockedStmt, err := p.db.PrepareContext(ctx, `
		SELECT 
			sequence,
			event_id,
			type,
			version,
			aggregate_id,
			payload,
			metadata,
			timestamp,
			attempt
		FROM outbox WHERE reserved_by = ?
	`)
	if err != nil {
		return fmt.Errorf("publish create locked stmt: %w", err)
	}
	defer lockedStmt.Close()

	locked := func(ctx context.Context, ident string) ([]Event, error) {

		rows, err := lockedStmt.QueryContext(ctx, ident, ident, ident)
		if err != nil {
			return nil, fmt.Errorf("fetch locked: %w", err)
		}
		defer rows.Close()

		events := make([]Event, 0, 100)
		for rows.Next() {
			event := Event{}
			err := rows.Scan(
				&event.Sequence,
				&event.ID,
				&event.Type,
				&event.Version,
				&event.AggregateID,
				&event.Payload,
				&event.Metadata,
				&event.Timestamp,
			)
			if err != nil {
				return nil, fmt.Errorf("publish row scan: %w", err)
			}
			events = append(events, event)
		}

		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("process locked rows: %w", err)

		}

		return events, nil
	}

	markDelivered := func(ctx context.Context, events []Event) error {
		if len(events) == 0 {
			return nil
		}
		placeholders := strings.Join(slices.Repeat([]string{"?"}, len(events)), ",")
		sequenceIds, err := Fmap(events, func(e Event) (any, error) {
			return e.Sequence, nil
		})
		if err != nil {
			return err
		}
		if _, err := p.db.ExecContext(ctx, fmt.Sprintf(`
			UPDATE outbox SET
			STATUS = %d
			WHERE sequence IN (%s)
		`, OutboxStatusDelivered, placeholders), sequenceIds...); err != nil {
			return fmt.Errorf("mark delivered: %w", err)
		}
		return nil
	}

	markFailed := func(ctx context.Context, events []Event) error {
		if len(events) == 0 {
			return nil
		}

		updateErrorValuesQuery := fmt.Sprintf(`
			WITH updated(sequence, error, reserved_at, reserved_by) AS (VALUES
				%s
			)
			UPDATE outbox 
			SET
				error = updated.error,
				reserved_at = updated.reserved_at,
				reserved_by = updated_reserved_by
			FROM updated
			WHERE (outbox.sequence = updated.sequence)
		`, strings.Join(MustFmap(events, func(_ Event) (string, error) {
			return "(?, ?, NULL, NULL)", nil
		}), ","))

		updateErrorValues := make([]any, len(events)*2)
		for _, evt := range events {
			updateErrorValues = append(updateErrorValues, evt.Sequence, evt.Error)
		}

		if _, err := p.db.ExecContext(ctx, updateErrorValuesQuery, updateErrorValues...); err != nil {
			return err
		}

		return nil
	}

	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("publish get hostname: %w", err)
	}

	pid := os.Getpid()

	var wg sync.WaitGroup
	for i := range p.workers {
		ident := fmt.Sprintf("h%s:p%d:wid%d", hostname, pid, i)
		wg.Add(1)
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					if err := lock(ctx, ident); err != nil {
						p.logger.Error(err.Error())
						break
					}

					events, err := locked(ctx, ident)
					if err != nil {
						p.logger.Error(err.Error())
						break
					}

					failedEvents := []Event{}
					deliveredEvents := []Event{}

					for _, event := range events {
						serde, err := NewSerdeJSONEvent(event)
						if err != nil {
							event.Error = err
							failedEvents = append(failedEvents, event)
							continue
						}
						body, err := json.Marshal(serde)
						if err != nil {
							event.Error = err
							failedEvents = append(failedEvents, event)
							continue
						}
						if err := ch.PublishWithContext(
							ctx,
							"task_manager",
							fmt.Sprintf("task_manager.%s.%s", event.Type, event.AggregateID),
							false, // mandatory
							false, // immediate
							// TODO: setup schema registry
							amqp.Publishing{
								ContentType: "application/json",
								Body:        body,
							},
						); err != nil {
							failedEvents = append(failedEvents, event)
							continue
						}

						deliveredEvents = append(deliveredEvents, event)
					}

					if err := markDelivered(ctx, deliveredEvents); err != nil {
						p.logger.Error(err.Error())
						break
					}

					if err := markFailed(ctx, failedEvents); err != nil {
						p.logger.Error(err.Error())
						break
					}
				}
			}
		}()
	}

	wg.Wait()
	return nil
}
