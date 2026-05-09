package consumer

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/tmcrypted/berusmenu/notification-service/internal/config"
	"github.com/tmcrypted/berusmenu/notification-service/internal/dto"
	notificationservice "github.com/tmcrypted/berusmenu/notification-service/internal/service"
	eventsv1 "github.com/tmcrypted/berusmenu/proto/gen/go/events/v1"
	"google.golang.org/protobuf/proto"
)

func TestProcessMessageCommitDecisions(t *testing.T) {
	cfg := config.Config{
		VerificationTopic: "auth.email_verification_requested",
		ResponseTopic:     "matching.response_created",
		RetryAttempts:     2,
		RetryBackoff:      time.Millisecond,
	}

	t.Run("success commits", func(t *testing.T) {
		sender := &fakeSender{}
		c := NewWithReader(cfg, notificationservice.New(sender), nil)
		msg := kafka.Message{Topic: cfg.VerificationTopic, Value: mustMarshal(t, &eventsv1.EmailVerificationRequested{
			EventId: "event-1",
			UserId:  7,
			Email:   "worker@example.com",
			Code:    "123456",
		})}

		if !c.processMessage(context.Background(), msg) {
			t.Fatal("expected commit")
		}
		if sender.emailVerificationCalls != 1 {
			t.Fatalf("email calls = %d, want 1", sender.emailVerificationCalls)
		}
	})

	t.Run("malformed message commits", func(t *testing.T) {
		c := NewWithReader(cfg, notificationservice.New(&fakeSender{}), nil)
		msg := kafka.Message{Topic: cfg.VerificationTopic, Value: []byte("not-protobuf")}

		if !c.processMessage(context.Background(), msg) {
			t.Fatal("expected malformed message to commit")
		}
	})

	t.Run("sender failure does not commit after retries", func(t *testing.T) {
		senderErr := errors.New("smtp down")
		sender := &fakeSender{err: senderErr}
		c := NewWithReader(cfg, notificationservice.New(sender), nil)
		msg := kafka.Message{Topic: cfg.ResponseTopic, Value: mustMarshal(t, &eventsv1.ResponseCreated{
			EventId:      "event-2",
			ResponseId:   11,
			VacancyId:    22,
			OwnerUserId:  33,
			OwnerEmail:   "owner@example.com",
			WorkerUserId: 44,
		})}

		if c.processMessage(context.Background(), msg) {
			t.Fatal("expected no commit")
		}
		if sender.responseCreatedCalls != cfg.RetryAttempts {
			t.Fatalf("response calls = %d, want %d", sender.responseCreatedCalls, cfg.RetryAttempts)
		}
	})

	t.Run("duplicate event commits without sending", func(t *testing.T) {
		sender := &fakeSender{}
		dedup := &fakeDedup{reserved: false}
		c := NewWithReader(cfg, notificationservice.New(sender), nil)
		c.dedup = dedup
		msg := kafka.Message{Topic: cfg.VerificationTopic, Value: mustMarshal(t, &eventsv1.EmailVerificationRequested{
			EventId: "event-3",
			UserId:  7,
			Email:   "worker@example.com",
			Code:    "123456",
		})}

		if !c.processMessage(context.Background(), msg) {
			t.Fatal("expected duplicate message to commit")
		}
		if sender.emailVerificationCalls != 0 {
			t.Fatalf("email calls = %d, want 0", sender.emailVerificationCalls)
		}
	})
}

type fakeSender struct {
	err                    error
	emailVerificationCalls int
	responseCreatedCalls   int
}

func (s *fakeSender) SendEmailVerification(ctx context.Context, notification dto.EmailVerificationRequested) error {
	_ = ctx
	s.emailVerificationCalls++
	return s.err
}

func (s *fakeSender) SendResponseCreated(ctx context.Context, notification dto.ResponseCreated) error {
	_ = ctx
	s.responseCreatedCalls++
	return s.err
}

type fakeDedup struct {
	reserved bool
}

func (d *fakeDedup) ReserveEvent(ctx context.Context, eventID string, eventType string, timeout time.Duration) (bool, error) {
	_, _, _, _ = ctx, eventID, eventType, timeout
	return d.reserved, nil
}

func (d *fakeDedup) MarkEventProcessed(ctx context.Context, eventID string) error {
	_, _ = ctx, eventID
	return nil
}

func (d *fakeDedup) ReleaseEvent(ctx context.Context, eventID string) error {
	_, _ = ctx, eventID
	return nil
}

func mustMarshal(t *testing.T, msg proto.Message) []byte {
	t.Helper()
	data, err := proto.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
