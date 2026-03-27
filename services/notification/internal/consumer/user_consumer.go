package consumer

import (
	"context"
	"encoding/json"
	"log"
	"notification/config"

	"github.com/segmentio/kafka-go"
)

type UserEvent struct {
	EventType string `json:"event_type"`
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	Timestamp string `json:"timestamp"`
}

type UserConsumer struct {
	reader *kafka.Reader
}

func NewUserConsumer(cfg *config.Config) *UserConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: cfg.Kafka.Brokers,
		Topic:   cfg.Kafka.Topic,
		GroupID: cfg.Kafka.GroupID,
	})

	return &UserConsumer{
		reader: reader,
	}
}

func (c *UserConsumer) Start(ctx context.Context) {
	log.Println("Starting user consumer...")

	for {
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		var event UserEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("Error unmarshaling event: %v", err)
			continue
		}

		// Обработка события
		log.Printf("📧 New user registered: %s (%s)", event.Email, event.Username)
		// Здесь можно отправить приветственное письмо
	}
}

func (c *UserConsumer) Close() error {
	return c.reader.Close()
}
