package events

import (
	"context"

	"github.com/segmentio/kafka-go"
	"github.com/tmcrypted/berusmenu/matching-service/internal/config"
	"github.com/tmcrypted/berusmenu/matching-service/internal/store"
	eventsv1 "github.com/tmcrypted/berusmenu/proto/gen/go/events/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const ResponseCreatedType = "berusmenu.events.v1.ResponseCreated"

type Builder struct {
	cfg config.Config
}

type KafkaPublisher struct {
	writer *kafka.Writer
}

func NewBuilder(cfg config.Config) *Builder {
	return &Builder{cfg: cfg}
}

func NewKafkaPublisher(writer *kafka.Writer) *KafkaPublisher {
	return &KafkaPublisher{writer: writer}
}

func (b *Builder) ResponseCreated(responseID, vacancyID, ownerUserID int64, ownerEmail string, workerUserID int64, vacancyTitle string, shiftStartAt string, payAmount int64) (store.OutboxEventInput, error) {
	eventID, err := store.NewEventID()
	if err != nil {
		return store.OutboxEventInput{}, err
	}
	event := &eventsv1.ResponseCreated{
		EventId:      eventID,
		ResponseId:   responseID,
		VacancyId:    vacancyID,
		OwnerUserId:  ownerUserID,
		OwnerEmail:   ownerEmail,
		WorkerUserId: workerUserID,
		VacancyTitle: vacancyTitle,
		ShiftDate:    shiftStartAt,
		PayAmount:    payAmount,
		OccurredAt:   timestamppb.Now(),
	}
	payload, err := proto.Marshal(event)
	if err != nil {
		return store.OutboxEventInput{}, err
	}
	return store.OutboxEventInput{
		EventID:   eventID,
		Topic:     b.cfg.ResponseTopic,
		EventType: ResponseCreatedType,
		EventKey:  ownerEmail,
		Payload:   payload,
	}, nil
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
