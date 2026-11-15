package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fdelivery_orders/internal/order"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
)

const (
	procStartBackoff      = time.Millisecond * 500
	procMaxRetries        = 10
	procMaxLimit          = 30 * time.Second
	procBackoffMultiplier = 2
	procAttemptTimeout    = 5 * time.Second

	// dlq
	dlqStartBackoff      = time.Millisecond * 500
	dlqMaxRetries        = 3
	dlqMaxLimit          = 3 * time.Second
	dlqAttemptTimeout    = 3 * time.Second
	dlqBackoffMultiplier = 2
)

type Consumer struct {
	reader       *kafka.Reader
	dlqWriter    *kafka.Writer
	log          *slog.Logger
	txRunner     TxRunner
	workflowRepo WorkflowRepo
	workflow     TemporalAdapter
}

func NewConsumer(brokers []string, log *slog.Logger, txRunner TxRunner, workflowRepo WorkflowRepo, workflow TemporalAdapter) *Consumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          "OrderCreated",
		GroupID:        "order-created-group-v1",
		StartOffset:    kafka.FirstOffset,
		CommitInterval: 0,
	})

	w := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    "OrderCreated-dlq",
		Balancer: &kafka.LeastBytes{},
	}

	return &Consumer{
		reader:       r,
		dlqWriter:    w,
		log:          log,
		txRunner:     txRunner,
		workflowRepo: workflowRepo,
		workflow:     workflow,
	}
}

func (c *Consumer) Start(ctx context.Context) {
	defer func() {
		if err := c.reader.Close(); err != nil {
			c.log.Error("failed to close kafka reader", "error", err)
		}

		if err := c.dlqWriter.Close(); err != nil {
			c.log.Error("failed to close dlq writer", "error", err)
		}
	}()

	for {
		if ctx.Err() != nil {
			return
		}

		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				c.log.Info("stopping kafka consumer")
				return
			}
			c.log.Error("failed to read message", "error", err)

			if !sleepWithCtx(ctx, time.Second) {
				return
			}

			continue
		}

		var event order.Event
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			c.log.Error("failed to unmarshal message", "error", err, "offset", msg.Offset, "raw", string(msg.Value))

			c.handleDeadLetter(ctx, msg, err)

			if ctx.Err() != nil {
				return
			}

			c.commit(ctx, msg)
			continue
		}

		if event.Type != order.EventOrderCreated {
			c.log.Error("invalid event type", "want", order.EventOrderCreated, "have", event.Type)

			c.commit(ctx, msg)

			continue
		}

		if err := c.processWithJitterRetries(ctx, event); err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}

			c.log.Error("all retry attempts exhausted, send to dlq",
				"error", err,
				"event_id", event.ID,
				"offset", msg.Offset,
			)

			c.handleDeadLetter(ctx, msg, err)

			if ctx.Err() != nil {
				return
			}
		}

		c.commit(ctx, msg)
	}
}

func (c *Consumer) handleEvent(ctx context.Context, event order.Event) error {
	attemptCtx, cancel := context.WithTimeout(ctx, procAttemptTimeout)
	defer cancel()

	return c.txRunner.WithinTx(attemptCtx, func(ctx context.Context, q order.Querier) error {
		if err := c.workflowRepo.Save(ctx, q, event.ID, event.AggregateID); err != nil {
			return err
		}

		err := c.workflow.StartWorkflow(ctx, event.ID)
		if err != nil {
			if errors.Is(err, order.ErrWorkflowIsAlreadyStarted) {
				return nil
			}

			return err
		}

		return nil
	})
}

func (c *Consumer) processWithJitterRetries(ctx context.Context, event order.Event) error {
	backoff := procStartBackoff
	var lastErr error

	for attempt := 1; attempt <= procMaxRetries; attempt++ {
		lastErr = c.handleEvent(ctx, event)
		if lastErr == nil {
			return nil
		}

		if errors.Is(lastErr, context.Canceled) {
			return lastErr
		}

		c.log.Warn("failed to handle event, retrying",
			"attempt", attempt,
			"max_retries", procMaxRetries,
			"backoff", backoff,
			"event_id", event.ID,
			"error", lastErr,
		)

		if attempt == procMaxRetries {
			break
		}

		sleep := rand.N(backoff + 1)

		if !sleepWithCtx(ctx, sleep) {
			return context.Canceled
		}

		if backoff > procMaxLimit/time.Duration(procBackoffMultiplier) {
			backoff = procMaxLimit
		} else {
			backoff *= time.Duration(procBackoffMultiplier)
		}

	}

	return fmt.Errorf("after %d attempts: %w", procMaxRetries, lastErr)
}

func (c *Consumer) commit(ctx context.Context, msg kafka.Message) {
	commitCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()

	if err := c.reader.CommitMessages(commitCtx, msg); err != nil {
		c.log.Error("failed to commit message", "error", err)
	}
}

func sleepWithCtx(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func (c *Consumer) handleDeadLetter(ctx context.Context, msg kafka.Message, err error) {
	if dlqErr := c.sendToDLQWithRetries(ctx, msg, err); dlqErr != nil {
		if ctx.Err() != nil {
			return
		}

		c.log.Error("CRITICAL: failed to write dlq, dumping raw message to logs",
			"dlq_error", dlqErr,
			"processing_error", err,
			"topic", msg.Topic,
			"partition", msg.Partition,
			"offset", msg.Offset,
			"key", string(msg.Key),
			"value", string(msg.Value),
		)
	}
}

func (c *Consumer) sendToDLQWithRetries(ctx context.Context, msg kafka.Message, err error) error {
	headers := make([]kafka.Header, 0, len(msg.Headers)+4)
	headers = append(headers, msg.Headers...)
	headers = append(headers,
		kafka.Header{Key: "x-dlq-error", Value: []byte(err.Error())},
		kafka.Header{Key: "x-dlq-original-topic", Value: []byte(msg.Topic)},
		kafka.Header{Key: "x-dlq-original-partition", Value: []byte(strconv.FormatInt(int64(msg.Partition), 10))},
		kafka.Header{Key: "x-dlq-original-offset", Value: []byte(strconv.FormatInt(msg.Offset, 10))},
	)

	dlqMsg := kafka.Message{
		Key:     msg.Key,
		Value:   msg.Value,
		Headers: headers,
	}

	backoff := dlqStartBackoff
	var lastErr error

	for attempt := 1; attempt <= dlqMaxRetries; attempt++ {
		attemptCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), dlqAttemptTimeout)
		lastErr = c.dlqWriter.WriteMessages(attemptCtx, dlqMsg)
		cancel()

		if lastErr == nil {
			return nil
		}

		c.log.Warn("failed to write message to DLQ, retrying",
			"attempt", attempt,
			"max_retries", dlqMaxRetries,
			"backoff", backoff,
			"offset", msg.Offset,
			"error", lastErr,
		)

		if attempt == dlqMaxRetries {
			break
		}

		if !sleepWithCtx(ctx, backoff) {
			return ctx.Err()
		}

		if backoff > dlqMaxLimit/time.Duration(dlqBackoffMultiplier) {
			backoff = dlqMaxLimit
		} else {
			backoff *= time.Duration(dlqBackoffMultiplier)
		}
	}

	return fmt.Errorf("dlq write failed after %d attempts: %w", dlqMaxRetries, lastErr)
}
