package messagebroker

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/virhanali/filmnesia/user-service/internal/config"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQPublisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	cfg     config.Config
}

func NewRabbitMQPublisher(cfg config.Config) (*RabbitMQPublisher, error) {
	log.Printf("Attempting to connect to RabbitMQ at %s", cfg.RabbitMQURL)
	conn, err := amqp.Dial(cfg.RabbitMQURL)
	if err != nil {
		log.Printf("Failed to connect to RabbitMQ: %v", err)
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Printf("Failed to open a channel: %v", err)
		conn.Close()
		return nil, err
	}

	log.Println("Successfully connected to RabbitMQ and opened a channel.")
	return &RabbitMQPublisher{
		conn:    conn,
		channel: ch,
		cfg:     cfg,
	}, nil
}

func (p *RabbitMQPublisher) PublishUserRegisteredEvent(ctx context.Context, exchangeName, routingKey string, eventData interface{}) error {
	err := p.channel.ExchangeDeclare(
		exchangeName,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Printf("Failed to declare an exchange '%s': %v", exchangeName, err)
		return err
	}
	log.Printf("Exchange '%s' declared successfully or already exists.", exchangeName)

	body, err := json.Marshal(eventData)
	if err != nil {
		log.Printf("Failed to marshal event data to JSON: %v", err)
		return err
	}

	err = p.channel.PublishWithContext(ctx,
		exchangeName,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			Body:         body,
		})
	if err != nil {
		log.Printf("Failed to publish a message to exchange '%s' with routing key '%s': %v", exchangeName, routingKey, err)
		return err
	}

	log.Printf("Message published to exchange '%s' with routing key '%s'. Body: %s", exchangeName, routingKey, string(body))
	return nil
}

func (p *RabbitMQPublisher) Close() {
	if p.channel != nil {
		err := p.channel.Close()
		if err != nil {
			log.Printf("Error closing RabbitMQ channel: %v", err)
		} else {
			log.Println("RabbitMQ channel closed.")
		}
	}
	if p.conn != nil {
		err := p.conn.Close()
		if err != nil {
			log.Printf("Error closing RabbitMQ connection: %v", err)
		} else {
			log.Println("RabbitMQ connection closed.")
		}
	}
}
