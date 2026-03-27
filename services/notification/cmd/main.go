package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"notification/config"
	"notification/internal/consumer"
)

func main() {
	cfg := &config.Config{
		Kafka: config.KafkaConfig{
			Brokers: []string{"localhost:9092"},
			Topic:   "user-events",
			GroupID: "notification-group",
		},
	}

	userConsumer := consumer.NewUserConsumer(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Запускаем consumer в горутине
	go userConsumer.Start(ctx)

	// Ждем сигнала завершения
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")
	cancel()
	time.Sleep(2 * time.Second)

	if err := userConsumer.Close(); err != nil {
		log.Printf("Error closing consumer: %v", err)
	}

	log.Println("Notification service stopped")
}
