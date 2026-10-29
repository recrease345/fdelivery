package order

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Querier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type TxRunner interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context, q Querier) error) error
}

type OrderRepository interface {
	Save(ctx context.Context, q Querier, o *Order) error
	GetByID(ctx context.Context, id uuid.UUID) (*Order, error)
	UpdateStatus(ctx context.Context, q Querier, id uuid.UUID, from, to Status) error
}

type IdempotencyRepository interface {
	Save(ctx context.Context, q Querier, key string, orderID uuid.UUID, responseBody json.RawMessage, now time.Time) error
	GetResponse(ctx context.Context, key string) (uuid.UUID, json.RawMessage, error)
}

type OutboxRepository interface {
	Save(ctx context.Context, q Querier, ev Event) error
}

type EventPublisher interface { 
	Publish(ctx context.Context, ev Event) error
}
