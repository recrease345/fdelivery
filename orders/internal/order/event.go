package order

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

type EventType string

const (
	EventOrderCreated   EventType = "OrderCreated"
	EventOrderPaid      EventType = "OrderPaid"
	EventOrderCancelled EventType = "OrderCancelled"
)

type Event struct {
	ID          uuid.UUID
	Type        EventType
	AggregateID uuid.UUID
	Payload     json.RawMessage
	CreatedAt   time.Time
}

func NewEvent(id uuid.UUID, eventType EventType, aggregateID uuid.UUID, payload json.RawMessage, now time.Time) (*Event, error) {
	if id == uuid.Nil {
		return nil, ErrInvalidEventID
	}

	if aggregateID == uuid.Nil {
		return nil, ErrInvalidEventAggregateID
	}

	switch eventType {
	case EventOrderCreated, EventOrderPaid, EventOrderCancelled:
	default:
		return nil, ErrInvalidEventType
	}

	return &Event{
		ID:          id,
		Type:        eventType,
		AggregateID: aggregateID,
		Payload:     payload,
		CreatedAt:   now,
	}, nil
}

var (
	ErrDuplicateIdempotencyKey = errors.New("duplicete idempotency key")
	ErrKeyNotFound             = errors.New("idempotency key not found")
	ErrInvalidEventID          = errors.New("invalid event id")
	ErrInvalidEventAggregateID = errors.New("invalid aggregate id")
	ErrInvalidEventType        = errors.New("invalid event type")
	ErrEventNotMapped          = errors.New("event not mapped")
)

var eventsByStatus = map[Status]EventType{
	Paid:      EventOrderPaid,
	Cancelled: EventOrderCancelled,
}
