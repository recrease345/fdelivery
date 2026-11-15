package kafka

import (
	"context"
	"fdelivery_orders/internal/order"

	"github.com/google/uuid"
)

type WorkflowStarter interface {
	Start(ctx context.Context, orderID uuid.UUID) error
}

type WorkflowRepo interface {
	Save(ctx context.Context, q order.Querier, eventID uuid.UUID, orderID uuid.UUID) error
}

type TemporalAdapter interface {
	StartWorkflow(ctx context.Context, orderID uuid.UUID) error
}

type TxRunner interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context, q order.Querier) error) error
}
