package order

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

var ErrEmptyKey = errors.New("empty idempotency key")

type Service struct {
	orders   OrderRepository
	keys     IdempotencyRepository
	outbox   OutboxRepository
	txRunner TxRunner
	pub      EventPublisher
	log      *slog.Logger
}

type CreateCommand struct {
	UserID, RestarauntID uuid.UUID
	Items                []Item
	DeliveryAddress      string
	IdempotencyKey       string
}

type CreateResult struct {
	OrderID      uuid.UUID
	ResponseBody json.RawMessage
	Replay       bool
}

func (s *Service) Create(ctx context.Context, cmd CreateCommand) (*CreateResult, error) {
	if cmd.IdempotencyKey == "" {
		return nil, ErrEmptyKey
	}

	now := time.Now()

	o, err := NewOrder(cmd.UserID, cmd.RestarauntID, cmd.Items, cmd.DeliveryAddress, now)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	responseBody, err := json.Marshal(map[string]any{
		"order_id": o.id,
		"status":   o.status,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response body: %w", err)
	}

	payload, err := json.Marshal(map[string]any{
		"order_id":      o.id,
		"user_id":       o.userID,
		"restaurant_id": o.restaurantID,
		"total_price":   o.totalPrice,
		"status":        o.status,
		"created_at":    o.createdAt,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event payload: %w", err)
	}

	event, err := NewEvent(uuid.New(), EventOrderCreated, o.id, payload, now)
	if err != nil {
		return nil, fmt.Errorf("failed to create event: %w", err)
	}

	err = s.txRunner.WithinTx(ctx, func(ctx context.Context, q Querier) error {
		if err := s.keys.Save(ctx, q, cmd.IdempotencyKey, o.id, responseBody, now); err != nil {
			return fmt.Errorf("failed to save idempotency key: %w", err)
		}

		if err := s.orders.Save(ctx, q, o); err != nil {
			return fmt.Errorf("failed to save order: %w", err)
		}

		if err := s.outbox.Save(ctx, q, *event); err != nil {
			return fmt.Errorf("failed to save outbox: %w", err)
		}

		return nil
	})
	if err != nil {
		if errors.Is(err, ErrDuplicateIdempotencyKey) {
			orderId, replayResponseBody, err := s.keys.GetResponse(ctx, cmd.IdempotencyKey)
			if err != nil {
				return nil, fmt.Errorf("failed to get idempotency key: %w", err)
			}

			return &CreateResult{
				OrderID:      orderId,
				ResponseBody: replayResponseBody,
				Replay:       true,
			}, nil
		}

		return nil, fmt.Errorf("failed to create order tx: %w", err)
	}

	if err := s.pub.Publish(ctx, *event); err != nil {
		s.log.Error("failed to publish event", "eventType", event.Type, "event", event, "error", err)
		return &CreateResult{
			OrderID:      o.id,
			ResponseBody: responseBody,
			Replay:       false,
		}, nil
	}

	return &CreateResult{
		OrderID:      o.id,
		ResponseBody: responseBody,
		Replay:       false,
	}, nil
}

func (s *Service) Transit(ctx context.Context, id uuid.UUID, from, to Status) error {
	if !CanTransitionTo(from, to) {
		return &InvalidTransitionError{From: from, To: to}
	}

	et, ok := eventsByStatus[to]

	if !ok {
		return fmt.Errorf("%w: %s", ErrEventNotMapped, to)
	}

	now := time.Now()

	payload, err := json.Marshal(map[string]any{
		"order_id":    id,
		"from":        from,
		"to":          to,
		"occurred_at": now,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	event, err := NewEvent(uuid.New(), et, id, payload, now)
	if err != nil {
		return fmt.Errorf("failed to create event: %w", err)
	}

	return s.txRunner.WithinTx(ctx, func(ctx context.Context, q Querier) error {
		if err := s.orders.UpdateStatus(ctx, q, id, from, to); err != nil {
			return fmt.Errorf("failed to update status: %w", err)
		}

		if err := s.outbox.Save(ctx, q, *event); err != nil {
			return fmt.Errorf("failed to save  outbox: %w", err)
		}

		return nil
	})
}

func (s *Service) Cancel(ctx context.Context, id, userID uuid.UUID) (Compensation, error) {
	o, err := s.orders.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrOrderNotFound) {
			return CompUnset, ErrOrderNotFound
		}

		return CompUnset, err
	}

	if o.UserID() != userID {
		return CompUnset, ErrOrderNotFound
	}

	from := o.Status()

	comp, err := o.cancel()
	if err != nil {
		return CompUnset, err
	}

	if err := s.Transit(ctx, id, from, o.Status()); err != nil {
		return CompUnset, fmt.Errorf("failed to trasit order status: %w", err)
	}

	return comp, nil
}
