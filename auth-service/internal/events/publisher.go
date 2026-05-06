package events

import (
	"context"
	"log"
	"strings"

	"github.com/segmentio/kafka-go"
	"gitlab.com/berusmenu/auth-service/internal/config"
	eventsv1 "gitlab.com/berusmenu/proto/gen/go/events/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Publisher struct {
	writer *kafka.Writer
	cfg    config.Config
}

func NewKafkaPublisher(writer *kafka.Writer, cfg config.Config) *Publisher {
	return &Publisher{writer: writer, cfg: cfg}
}

func (p *Publisher) EmailVerificationRequested(ctx context.Context, userID int64, email string, code string) error {
	event := &eventsv1.EmailVerificationRequested{
		UserId:     userID,
		Email:      email,
		Code:       code,
		OccurredAt: timestamppb.Now(),
	}
	data, err := proto.Marshal(event)
	if err != nil {
		return err
	}

	err = p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(email),
		Value: data,
	})
	if p.isDev() {
		log.Printf("email verification code for user_id=%d email=%s code=%s", userID, email, code)
		if err != nil {
			log.Printf("publish %s failed in dev/local: %v", p.cfg.VerificationTopic, err)
			return nil
		}
	}
	return err
}

func (p *Publisher) isDev() bool {
	env := strings.ToLower(strings.TrimSpace(p.cfg.AppEnv))
	return env == "" || env == "local" || env == "dev" || env == "development"
}
