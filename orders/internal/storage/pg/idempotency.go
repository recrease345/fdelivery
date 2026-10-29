package pg

import (
	"context"
	"encoding/json"
	"errors"
	"fdelivery_orders/internal/order"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type IdempotencyKeysRepo struct {
	pool *pgxpool.Pool
}

func (r *IdempotencyKeysRepo) Save(ctx context.Context, q order.Querier, key string, orderID uuid.UUID, responseBody json.RawMessage, now time.Time) error {
	query := `
	INSERT INTO orders.idempotency_keys (key, order_id, response_body, created_at)
	VALUES ($1, $2, $3, $4)`

	_, err := q.Exec(ctx, query, key, orderID, responseBody, now)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return order.ErrDuplicateIdempotencyKey
			}
		}

		return fmt.Errorf("failed to save idempotency key: %w", err)
	}

	return nil
}

func (r *IdempotencyKeysRepo) GetResponse(ctx context.Context, key string) (uuid.UUID, json.RawMessage, error) {
	query := `SELECT order_id, response_body
	FROM orders.idempotency_keys 
	WHERE key = $1`

	var responseBody json.RawMessage
	var orderID uuid.UUID

	err := r.pool.QueryRow(ctx, query, key).Scan(&responseBody, &orderID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, nil, order.ErrKeyNotFound
		}

		return uuid.Nil, nil, fmt.Errorf("failed to get response body: %w", err)
	}

	return orderID, responseBody, nil
}
