package pg

import (
	"context"
	"errors"
	"fdelivery_orders/internal/order"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type WorkflowStartsRepo struct{}

func (r WorkflowStartsRepo) Save(ctx context.Context, q order.Querier, eventID uuid.UUID, orderID uuid.UUID) error {
	query := `INSERT INTO workflow_starts (event_id, order_id)
	VALUES ($1, $2)`

	if _, err := q.Exec(ctx, query, eventID, orderID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "workflow_starts_pkey" {
				return order.ErrWorkflowIsAlreadyStarted
			}
		}

		return fmt.Errorf("save workflow start: %w", err)
	}

	return nil
}
