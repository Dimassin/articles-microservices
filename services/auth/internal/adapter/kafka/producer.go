package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

type UserEvent struct {
	EventType string `json:"event_type"`
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	Timestamp string `json:"timestamp"`
}

type EventProducer struct {
	writer *kafka.Writer
	topic  string
}

func NewEventProducer(brokers []string, topic string) *EventProducer {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}

	return &EventProducer{
		writer: writer,
		topic:  topic,
	}
}

func (p *EventProducer) PublishUserCreated(ctx context.Context, userID, email, username string) error {
	event := UserEvent{
		EventType: "user_created",
		UserID:    userID,
		Email:     email,
		Username:  username,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Key:   []byte(userID),
		Value: data,
	}

	err = p.writer.WriteMessages(ctx, msg)
	if err != nil {
		log.Printf("Failed to send Kafka message: %v", err)
		return err
	}

	log.Printf("📤 Event sent to Kafka: user_created %s (%s)", email, username)
	return nil
}

func (p *EventProducer) Close() error {
	return p.writer.Close()
}
