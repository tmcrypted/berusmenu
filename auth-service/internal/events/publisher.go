package events

import (
	"context"
	"log"
	"strings"

	"github.com/segmentio/kafka-go"
	"github.com/tmcrypted/berusmenu/auth-service/internal/config"
	"github.com/tmcrypted/berusmenu/auth-service/internal/store"
	eventsv1 "github.com/tmcrypted/berusmenu/proto/gen/go/events/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const EmailVerificationRequestedType = "berusmenu.events.v1.EmailVerificationRequested"

type Outbox interface {
	EnqueueOutboxEvent(ctx context.Context, input store.OutboxEventInput) error
}

type Publisher struct {
	outbox Outbox
	cfg    config.Config
}

type KafkaPublisher struct {
	writer *kafka.Writer
	cfg    config.Config
}

func NewOutboxPublisher(outbox Outbox, cfg config.Config) *Publisher {
	return &Publisher{outbox: outbox, cfg: cfg}
}

func NewKafkaPublisher(writer *kafka.Writer, cfg config.Config) *KafkaPublisher {
	return &KafkaPublisher{writer: writer, cfg: cfg}
}

func (p *Publisher) EmailVerificationRequested(ctx context.Context, userID int64, email string, code string) error {
	eventID, err := store.NewEventID()
	if err != nil {
		return err
	}
	event := &eventsv1.EmailVerificationRequested{
		EventId:    eventID,
		UserId:     userID,
		Email:      email,
		Code:       code,
		OccurredAt: timestamppb.Now(),
	}
	data, err := proto.Marshal(event)
	if err != nil {
		return err
	}

	if p.isDev() {
		log.Printf("email verification code for user_id=%d email=%s code=%s event_id=%s", userID, email, code, eventID)
	}

	return p.outbox.EnqueueOutboxEvent(ctx, store.OutboxEventInput{
		EventID:   eventID,
		Topic:     p.cfg.VerificationTopic,
		EventType: EmailVerificationRequestedType,
		EventKey:  email,
		Payload:   data,
	})
}

func (p *KafkaPublisher) Publish(ctx context.Context, event store.OutboxEvent) error {
	return p.writer.WriteMessages(ctx, kafka.Message{
		Topic: event.Topic,
		Key:   []byte(event.EventKey),
		Value: event.Payload,
		Headers: []kafka.Header{
			{Key: "event_id", Value: []byte(event.EventID)},
			{Key: "event_type", Value: []byte(event.EventType)},
		},
	})
}

func (p *Publisher) isDev() bool {
	env := strings.ToLower(strings.TrimSpace(p.cfg.AppEnv))
	return env == "" || env == "local" || env == "dev" || env == "development"
}
