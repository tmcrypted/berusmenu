package events

import (
	"context"
	"log"
	"time"

	"github.com/tmcrypted/berusmenu/auth-service/internal/config"
	"github.com/tmcrypted/berusmenu/auth-service/internal/store"
)

type OutboxStore interface {
	ClaimPendingOutboxEvents(ctx context.Context, limit int, claimTimeout time.Duration) ([]store.OutboxEvent, error)
	MarkOutboxEventPublished(ctx context.Context, id int64) error
	MarkOutboxEventFailed(ctx context.Context, id int64, cause error, retryBackoff time.Duration) error
}

type EventPublisher interface {
	Publish(ctx context.Context, event store.OutboxEvent) error
}

type OutboxWorker struct {
	store     OutboxStore
	publisher EventPublisher
	cfg       config.Config
}

func NewOutboxWorker(store OutboxStore, publisher EventPublisher, cfg config.Config) *OutboxWorker {
	return &OutboxWorker{store: store, publisher: publisher, cfg: cfg}
}

func (w *OutboxWorker) Run(ctx context.Context) {
	interval := w.cfg.OutboxInterval
	if interval <= 0 {
		interval = time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("auth outbox worker started interval=%s batch_size=%d", interval, w.batchSize())
	for {
		w.publishBatch(ctx)
		select {
		case <-ctx.Done():
			log.Println("auth outbox worker stopped")
			return
		case <-ticker.C:
		}
	}
}

func (w *OutboxWorker) publishBatch(ctx context.Context) {
	events, err := w.store.ClaimPendingOutboxEvents(ctx, w.batchSize(), w.claimTimeout())
	if err != nil {
		if ctx.Err() == nil {
			log.Printf("auth outbox claim failed: %v", err)
		}
		return
	}
	if len(events) == 0 {
		return
	}

	for _, event := range events {
		if err := w.publisher.Publish(ctx, event); err != nil {
			log.Printf("auth outbox publish failed event_id=%s topic=%s attempt=%d: %v", event.EventID, event.Topic, event.Attempts+1, err)
			if markErr := w.store.MarkOutboxEventFailed(ctx, event.ID, err, w.retryBackoff(event.Attempts+1)); markErr != nil && ctx.Err() == nil {
				log.Printf("auth outbox mark failed event_id=%s: %v", event.EventID, markErr)
			}
			continue
		}
		if err := w.store.MarkOutboxEventPublished(ctx, event.ID); err != nil {
			log.Printf("auth outbox mark published failed event_id=%s: %v", event.EventID, err)
			continue
		}
		log.Printf("auth outbox published event_id=%s topic=%s", event.EventID, event.Topic)
	}
}

func (w *OutboxWorker) batchSize() int {
	if w.cfg.OutboxBatchSize < 1 {
		return 50
	}
	return w.cfg.OutboxBatchSize
}

func (w *OutboxWorker) retryBackoff(attempt int) time.Duration {
	backoff := w.cfg.OutboxRetryBackoff
	if backoff <= 0 {
		backoff = time.Second
	}
	if attempt < 1 {
		attempt = 1
	}
	if attempt > 30 {
		attempt = 30
	}
	return time.Duration(attempt) * backoff
}

func (w *OutboxWorker) claimTimeout() time.Duration {
	if w.cfg.OutboxClaimTimeout <= 0 {
		return 5 * time.Minute
	}
	return w.cfg.OutboxClaimTimeout
}
