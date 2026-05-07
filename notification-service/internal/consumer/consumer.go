package consumer

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/tmcrypted/berusmenu/notification-service/internal/config"
	"github.com/tmcrypted/berusmenu/notification-service/internal/dto"
	notificationservice "github.com/tmcrypted/berusmenu/notification-service/internal/service"
	eventsv1 "github.com/tmcrypted/berusmenu/proto/gen/go/events/v1"
	"google.golang.org/protobuf/proto"
)

type Consumer struct {
	reader  messageReader
	cfg     config.Config
	service *notificationservice.Service
}

type messageReader interface {
	FetchMessage(ctx context.Context) (kafka.Message, error)
	CommitMessages(ctx context.Context, msgs ...kafka.Message) error
	Close() error
}

func New(cfg config.Config, service *notificationservice.Service) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     cfg.KafkaBrokers,
		GroupID:     cfg.ConsumerGroupID,
		GroupTopics: []string{cfg.VerificationTopic, cfg.ResponseTopic},
		MinBytes:    10e3, // 10KB
		MaxBytes:    10e6, // 10MB
	})

	return NewWithReader(cfg, service, reader)
}

func NewWithReader(cfg config.Config, service *notificationservice.Service, reader messageReader) *Consumer {
	return &Consumer{
		reader:  reader,
		cfg:     cfg,
		service: service,
	}
}

func (c *Consumer) Run(ctx context.Context) error {
	log.Printf("starting consumer on topics: %s, %s", c.cfg.VerificationTopic, c.cfg.ResponseTopic)

	for {
		m, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			log.Printf("failed to fetch message: %v", err)
			continue
		}

		if c.processMessage(ctx, m) {
			if err := c.reader.CommitMessages(ctx, m); err != nil {
				log.Printf("failed to commit message: %v", err)
			}
		}
	}
}

func (c *Consumer) processMessage(ctx context.Context, m kafka.Message) bool {
	var err error
	switch m.Topic {
	case c.cfg.VerificationTopic:
		err = c.handleEmailVerification(ctx, m)
	case c.cfg.ResponseTopic:
		err = c.handleResponseCreated(ctx, m)
	default:
		log.Printf("unknown notification topic %q at offset %d", m.Topic, m.Offset)
		return true
	}

	if err == nil {
		return true
	}
	if errors.Is(err, notificationservice.ErrInvalidNotification) {
		log.Printf("invalid notification at topic %s offset %d: %v", m.Topic, m.Offset, err)
		return true
	}

	log.Printf("failed to handle notification at topic %s offset %d after retries: %v", m.Topic, m.Offset, err)
	return false
}

func (c *Consumer) handleEmailVerification(ctx context.Context, m kafka.Message) error {
	var event eventsv1.EmailVerificationRequested
	if err := proto.Unmarshal(m.Value, &event); err != nil {
		log.Printf("failed to unmarshal message at offset %d: %v", m.Offset, err)
		return notificationservice.ErrInvalidNotification
	}

	return c.withRetry(ctx, func() error {
		return c.service.HandleEmailVerificationRequested(ctx, dto.EmailVerificationRequested{
			UserID: event.UserId,
			Email:  event.Email,
			Code:   event.Code,
		})
	})
}

func (c *Consumer) handleResponseCreated(ctx context.Context, m kafka.Message) error {
	var event eventsv1.ResponseCreated
	if err := proto.Unmarshal(m.Value, &event); err != nil {
		log.Printf("failed to unmarshal message at offset %d: %v", m.Offset, err)
		return notificationservice.ErrInvalidNotification
	}

	return c.withRetry(ctx, func() error {
		return c.service.HandleResponseCreated(ctx, dto.ResponseCreated{
			ResponseID:   event.ResponseId,
			VacancyID:    event.VacancyId,
			OwnerUserID:  event.OwnerUserId,
			OwnerEmail:   event.OwnerEmail,
			WorkerUserID: event.WorkerUserId,
			VacancyTitle: event.VacancyTitle,
			ShiftDate:    event.ShiftDate,
			PayAmount:    event.PayAmount,
		})
	})
}

func (c *Consumer) withRetry(ctx context.Context, fn func() error) error {
	attempts := c.cfg.RetryAttempts
	if attempts < 1 {
		attempts = 1
	}
	backoff := c.cfg.RetryBackoff
	if backoff <= 0 {
		backoff = time.Second
	}

	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		err := fn()
		if err == nil || errors.Is(err, notificationservice.ErrInvalidNotification) {
			return err
		}
		lastErr = err
		if attempt == attempts {
			break
		}

		timer := time.NewTimer(backoff)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
	return lastErr
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
