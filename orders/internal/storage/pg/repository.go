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

type OrderRepo struct {
	pool *pgxpool.Pool
}

func (r *OrderRepo) Save(ctx context.Context, q order.Querier, o *order.Order) error {
	query := `
	INSERT INTO orders.orders
	(id, user_id, restaurant_id, items, total_price, delivery_address, status, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8)
	`

	items, err := json.Marshal(o.Items())
	if err != nil {
		return fmt.Errorf("failed to marshal items: %w", err)
	}

	_, err = q.Exec(ctx, query,
		o.ID(),
		o.UserID(),
		o.RestaurantID(),
		items,
		o.TotalPrice(),
		o.DeliveryAddress(),
		o.Status(),
		o.CreatedAt(),
	)
	if err != nil {
		return fmt.Errorf("failed to save order: %w", err)
	}

	return nil
}

func (r *OrderRepo) GetByID(ctx context.Context, id uuid.UUID) (*order.Order, error) {
	query := `
	SELECT id, user_id, restaurant_id, items, total_price, delivery_address, status, created_at
	FROM orders.orders WHERE id = $1`

	var (
		orderID         uuid.UUID
		userID          uuid.UUID
		restaurantID    uuid.UUID
		itemsRaw        []byte
		totalPrice      int64
		deliveryAddress string
		status          order.Status
		createdAt       time.Time
	)

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&orderID,
		&userID,
		&restaurantID,
		&itemsRaw,
		&totalPrice,
		&deliveryAddress,
		&status,
		&createdAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, order.ErrOrderNotFound
		}

		return nil, fmt.Errorf("failed to get order by id: %w", err)
	}

	var items []order.Item
	if err := json.Unmarshal(itemsRaw, &items); err != nil {
		return nil, fmt.Errorf("failed to unmarshal order items: %w", err)
	}

	o, err := order.Restore(orderID, userID, restaurantID, totalPrice, items, deliveryAddress, createdAt, status)
	if err != nil {
		return nil, fmt.Errorf("failed to restore order: %w", err)
	}

	return o, nil
}

func (r *OrderRepo) UpdateStatus(ctx context.Context, q order.Querier, id uuid.UUID, from order.Status, to order.Status) error {
	query := `
	UPDATE orders.orders
	SET status = $1, updated_at = now()
	WHERE id = $2 AND status = $3`

	tag, err := q.Exec(ctx, query, to, id, from)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			if pgErr.Code == "23514" {
				return order.ErrInvalidStatus
			}
		}

		return fmt.Errorf("failed to update status(%s -> %s): %w", from, to, err)
	}

	if tag.RowsAffected() == 0 {
		var current order.Status
		err := q.QueryRow(ctx,
			`SELECT status FROM orders.orders WHERE id = $1`, id).Scan(&current)

		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return order.ErrOrderNotFound
		case err != nil:
			return fmt.Errorf("recheck status: %w", err)
		}

		return &order.InvalidTransitionError{From: current, To: to}
	}

	return nil
}
