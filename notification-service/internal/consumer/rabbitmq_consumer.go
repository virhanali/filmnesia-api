package consumer

import (
	"context"
	"encoding/json"
	"log"

	"github.com/virhanali/filmnesia/notification-service/internal/config"
	"github.com/virhanali/filmnesia/notification-service/internal/domain"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQConsumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	cfg     config.Config
}

func NewRabbitMQConsumer(cfg config.Config) (*RabbitMQConsumer, error) {
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

	return &RabbitMQConsumer{
		conn:    conn,
		channel: ch,
		cfg:     cfg,
	}, nil
}

func (c *RabbitMQConsumer) SetupAndConsumeUserRegisteredEvents(ctx context.Context, exchangeName, queueName, routingKey string) error {
	err := c.channel.ExchangeDeclare(
		exchangeName,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Printf("Failed to declare exchange '%s': %v", exchangeName, err)
		return err
	}
	log.Printf("Exchange '%s' declared successfully or already exists.", exchangeName)

	q, err := c.channel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Printf("Failed to declare queue '%s': %v", queueName, err)
		return err
	}
	log.Printf("Queue '%s' declared successfully or already exists.", q.Name)

	err = c.channel.QueueBind(
		q.Name,
		routingKey,
		exchangeName,
		false,
		nil,
	)
	if err != nil {
		log.Printf("Failed to bind queue '%s' to exchange '%s' with routing key '%s': %v", q.Name, exchangeName, routingKey, err)
		return err
	}
	log.Printf("Queue '%s' bound to exchange '%s' with routing key '%s'.", q.Name, exchangeName, routingKey)

	msgs, err := c.channel.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Printf("Failed to register a consumer for queue '%s': %v", q.Name, err)
		return err
	}

	log.Printf("Waiting for messages on queue '%s'. To exit press CTRL+C", q.Name)

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("Consumer context cancelled. Shutting down message processing.")
				c.Close()
				return
			case d, ok := <-msgs:
				if !ok {
					log.Println("RabbitMQ messages channel closed. Consumer stopping.")
					return
				}

				log.Printf("Received a message from queue '%s': %s", q.Name, d.Body)

				var event domain.UserRegisteredEvent
				if err := json.Unmarshal(d.Body, &event); err != nil {
					log.Printf("ERROR: Failed to unmarshal message body: %v. Body: %s", err, d.Body)
					if errNack := d.Nack(false, false); errNack != nil {
						log.Printf("ERROR: Failed to Nack unparseable message: %v", errNack)
					}
					continue
				}

				log.Printf("INFO: Processing UserRegisteredEvent - Sending welcome email to UserID: %s, Email: %s, Username: %s",
					event.UserID, event.Email, event.Username)

				if errAck := d.Ack(false); errAck != nil {
					log.Printf("ERROR: Failed to acknowledge message: %v", errAck)
				} else {
					log.Printf("Message acknowledged for UserID: %s", event.UserID)
				}
			}
		}
	}()

	return nil
}

func (c *RabbitMQConsumer) Close() {
	log.Println("Closing RabbitMQ consumer channel and connection...")
	if c.channel != nil {
		if err := c.channel.Close(); err != nil {
			log.Printf("Error closing RabbitMQ channel: %v", err)
		} else {
			log.Println("RabbitMQ channel closed.")
		}
		c.channel = nil
	}
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			log.Printf("Error closing RabbitMQ connection: %v", err)
		} else {
			log.Println("RabbitMQ connection closed.")
		}
		c.conn = nil
	}
}
