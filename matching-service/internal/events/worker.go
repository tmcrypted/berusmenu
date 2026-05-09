package events

import (
	"context"
	"log"
	"time"

	"github.com/tmcrypted/berusmenu/matching-service/internal/config"
	"github.com/tmcrypted/berusmenu/matching-service/internal/store"
)

type OutboxStore interface {
	ClaimPendingOutboxEvents(ctx context.Context, limit int, claimTimeout time.Duration) ([]store.OutboxEvent, error)
	MarkOutboxEventPublished(ctx context.Context, id int64) error
	MarkOutboxEventFailed(ctx context.Context, id int64, cause error, retryBackoff time.Duration) error
}

type EventPublisher interface {
	Publish(ctx context.Context, event store.OutboxEvent) error
}

type Worker struct {
	store     OutboxStore
	publisher EventPublisher
	cfg       config.Config
}

func NewWorker(store OutboxStore, publisher EventPublisher, cfg config.Config) *Worker {
	return &Worker{store: store, publisher: publisher, cfg: cfg}
}

func (w *Worker) Run(ctx context.Context) {
	interval := w.cfg.OutboxPollInterval
	if interval <= 0 {
		interval = time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	log.Printf("matching outbox worker started interval=%s batch_size=%d", interval, w.cfg.OutboxBatchSize)
	for {
		select {
		case <-ctx.Done():
			log.Println("matching outbox worker stopped")
			return
		case <-ticker.C:
			w.tick(ctx)
		}
	}
}

func (w *Worker) tick(ctx context.Context) {
	events, err := w.store.ClaimPendingOutboxEvents(ctx, w.cfg.OutboxBatchSize, w.cfg.OutboxClaimTimeout)
	if err != nil {
		log.Printf("matching outbox claim failed: %v", err)
		return
	}
	for _, event := range events {
		if err := w.publisher.Publish(ctx, event); err != nil {
			log.Printf("matching outbox publish failed event_id=%s: %v", event.EventID, err)
			if markErr := w.store.MarkOutboxEventFailed(ctx, event.ID, err, w.cfg.OutboxRetryBackoff); markErr != nil {
				log.Printf("matching outbox mark failed event_id=%s: %v", event.EventID, markErr)
			}
			continue
		}
		if err := w.store.MarkOutboxEventPublished(ctx, event.ID); err != nil {
			log.Printf("matching outbox mark published failed event_id=%s: %v", event.EventID, err)
		}
	}
}
