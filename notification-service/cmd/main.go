package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/virhanali/filmnesia/notification-service/internal/config"
	"github.com/virhanali/filmnesia/notification-service/internal/consumer"
)

const (
	UserEventsExchange       = "user_events"
	UserRegisteredQueue      = "user.registered.notifications.queue"
	UserRegisteredRoutingKey = "user.registered"
)

func main() {
	cfg, err := config.LoadConfig("../")
	if err != nil {
		log.Fatalf("FATAL: Failed to load configuration: %v", err)
	}

	if cfg.RabbitMQURL == "" {
		log.Fatal("FATAL: RABBITMQ_URL is not set in configuration.")
	}

	mqConsumer, err := consumer.NewRabbitMQConsumer(cfg)
	if err != nil {
		log.Fatalf("FATAL: Failed to initialize RabbitMQ consumer: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = mqConsumer.SetupAndConsumeUserRegisteredEvents(
		ctx,
		UserEventsExchange,
		UserRegisteredQueue,
		UserRegisteredRoutingKey,
	)
	if err != nil {
		log.Fatalf("FATAL: Failed to setup or start consuming user registered events: %v", err)
	}
	err = mqConsumer.SetupAndConsumeUserRegisteredEvents(
		ctx,
		UserEventsExchange,
		UserRegisteredQueue,
		UserRegisteredRoutingKey,
	)
	if err != nil {
		log.Fatalf("FATAL: Failed to setup or start consuming user registered events: %v", err)
	}

	log.Println("Notification Service is running and waiting for messages...")
	log.Println("Press CTRL+C to exit.")
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("INFO: Notification Service shutting down...")

	cancel()

	log.Println("INFO: Notification Service shutdown complete.")
}
