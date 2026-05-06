package consumer

import (
	"context"
	"log"

	"github.com/segmentio/kafka-go"
	"gitlab.com/berusmenu/notification-service/internal/config"
	eventsv1 "gitlab.com/berusmenu/proto/gen/go/events/v1"
	"google.golang.org/protobuf/proto"
)

type Consumer struct {
	reader *kafka.Reader
	cfg    config.Config
}

func New(cfg config.Config) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  cfg.KafkaBrokers,
		GroupID:  cfg.ConsumerGroupID,
		Topic:    cfg.VerificationTopic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	return &Consumer{
		reader: reader,
		cfg:    cfg,
	}
}

func (c *Consumer) Run(ctx context.Context) error {
	log.Printf("starting consumer on topic: %s", c.cfg.VerificationTopic)

	for {
		m, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil // context canceled
			}
			log.Printf("failed to fetch message: %v", err)
			continue
		}

		c.processMessage(m)

		if err := c.reader.CommitMessages(ctx, m); err != nil {
			log.Printf("failed to commit message: %v", err)
		}
	}
}

func (c *Consumer) processMessage(m kafka.Message) {
	var event eventsv1.EmailVerificationRequested
	if err := proto.Unmarshal(m.Value, &event); err != nil {
		log.Printf("failed to unmarshal message at offset %d: %v", m.Offset, err)
		return
	}

	// Заглушка вместо реальной отправки email
	log.Printf("==================================================")
	log.Printf("[EMAIL STUB] To: %s", event.Email)
	log.Printf("[EMAIL STUB] Verification Code: %s", event.Code)
	log.Printf("[EMAIL STUB] User ID: %d", event.UserId)
	log.Printf("==================================================")
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
