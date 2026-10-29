package pg

import (
	"context"
	"fdelivery_orders/internal/order"
	"fmt"
)

type OutboxRepo struct{}

func (r *OutboxRepo) Save(ctx context.Context, q order.Querier, ev order.Event) error {
	query := `INSERT INTO orders.outbox
	(event_id, event_type, aggregate_id, payload, created_at)
	VALUES ($1, $2, $3, $4, $5)`

	_, err := q.Exec(ctx, query, ev.ID, ev.Type, ev.AggregateID, ev.Payload, ev.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to save outbox: %w", err)
	}

	return nil
}
