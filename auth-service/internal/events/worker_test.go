package events

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/tmcrypted/berusmenu/auth-service/internal/config"
	"github.com/tmcrypted/berusmenu/auth-service/internal/store"
)

func TestOutboxWorkerPublishesAndMarksPublished(t *testing.T) {
	repo := &fakeOutboxStore{
		events: []store.OutboxEvent{{ID: 1, EventID: "event-1", Topic: "topic"}},
	}
	publisher := &fakeEventPublisher{}
	worker := NewOutboxWorker(repo, publisher, config.Config{OutboxBatchSize: 50})

	worker.publishBatch(context.Background())

	if publisher.calls != 1 {
		t.Fatalf("publish calls = %d, want 1", publisher.calls)
	}
	if repo.publishedID != 1 {
		t.Fatalf("published id = %d, want 1", repo.publishedID)
	}
	if repo.failedID != 0 {
		t.Fatalf("failed id = %d, want 0", repo.failedID)
	}
}

func TestOutboxWorkerRetriesFailedPublish(t *testing.T) {
	publishErr := errors.New("kafka down")
	repo := &fakeOutboxStore{
		events: []store.OutboxEvent{{ID: 2, EventID: "event-2", Topic: "topic", Attempts: 1}},
	}
	publisher := &fakeEventPublisher{err: publishErr}
	worker := NewOutboxWorker(repo, publisher, config.Config{
		OutboxBatchSize:    50,
		OutboxRetryBackoff: time.Second,
	})

	worker.publishBatch(context.Background())

	if publisher.calls != 1 {
		t.Fatalf("publish calls = %d, want 1", publisher.calls)
	}
	if repo.publishedID != 0 {
		t.Fatalf("published id = %d, want 0", repo.publishedID)
	}
	if repo.failedID != 2 {
		t.Fatalf("failed id = %d, want 2", repo.failedID)
	}
	if repo.failureErr == nil || !errors.Is(repo.failureErr, publishErr) {
		t.Fatalf("failure err = %v, want %v", repo.failureErr, publishErr)
	}
	if repo.failureBackoff != 2*time.Second {
		t.Fatalf("failure backoff = %s, want 2s", repo.failureBackoff)
	}
}

type fakeOutboxStore struct {
	events         []store.OutboxEvent
	publishedID    int64
	failedID       int64
	failureErr     error
	failureBackoff time.Duration
}

func (s *fakeOutboxStore) ClaimPendingOutboxEvents(ctx context.Context, limit int, claimTimeout time.Duration) ([]store.OutboxEvent, error) {
	_, _, _ = ctx, limit, claimTimeout
	return s.events, nil
}

func (s *fakeOutboxStore) MarkOutboxEventPublished(ctx context.Context, id int64) error {
	_ = ctx
	s.publishedID = id
	return nil
}

func (s *fakeOutboxStore) MarkOutboxEventFailed(ctx context.Context, id int64, cause error, retryBackoff time.Duration) error {
	_ = ctx
	s.failedID = id
	s.failureErr = cause
	s.failureBackoff = retryBackoff
	return nil
}

type fakeEventPublisher struct {
	err   error
	calls int
}

func (p *fakeEventPublisher) Publish(ctx context.Context, event store.OutboxEvent) error {
	_, _ = ctx, event
	p.calls++
	return p.err
}
