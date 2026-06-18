package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"articleswap-scalable/services/internal/config"
)

// Broker wraps RabbitMQ connection and channel for publish/consume operations.
type Broker struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	exchange string
	cfg      *config.Config
}

// NewBroker connects to RabbitMQ, declares the exchange and all queues,
// and binds them with appropriate routing keys.
func NewBroker(cfg *config.Config) (*Broker, error) {
	conn, err := amqp.Dial(cfg.RabbitMQURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare topic exchange.
	if err := ch.ExchangeDeclare(
		cfg.RabbitMQExchange,
		"topic",
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,
	); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Declare and bind queues.
	queues := map[string]string{
		cfg.QueueSubmitted:  "article.submitted",
		cfg.QueueStemming:   "article.stemming",
		cfg.QueueWordcloud:  "article.wordcloud",
		cfg.QueueAggregator: "article.aggregator",
		cfg.QueueFailed:     "article.failed",
	}

	for queueName, routingKey := range queues {
		if _, err := ch.QueueDeclare(
			queueName,
			true,  // durable
			false, // auto-delete
			false, // exclusive
			false, // no-wait
			nil,
		); err != nil {
			ch.Close()
			conn.Close()
			return nil, fmt.Errorf("failed to declare queue %s: %w", queueName, err)
		}

		if err := ch.QueueBind(
			queueName,
			routingKey,
			cfg.RabbitMQExchange,
			false,
			nil,
		); err != nil {
			ch.Close()
			conn.Close()
			return nil, fmt.Errorf("failed to bind queue %s: %w", queueName, err)
		}
	}

	// Set prefetch count for fair dispatch among workers.
	if err := ch.Qos(cfg.WorkerConcurrency, 0, false); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to set QoS: %w", err)
	}

	log.Printf("[rabbitmq] connected — exchange=%s queues=%d",
		cfg.RabbitMQExchange, len(queues))

	return &Broker{
		conn:     conn,
		channel:  ch,
		exchange: cfg.RabbitMQExchange,
		cfg:      cfg,
	}, nil
}

// Publish sends a JSON-encoded message to the exchange with the given routing key.
func (b *Broker) Publish(ctx context.Context, routingKey string, body interface{}) error {
	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return b.channel.PublishWithContext(ctx,
		b.exchange,
		routingKey,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         data,
			Timestamp:    time.Now(),
		},
	)
}

// Consume returns a channel of deliveries for the specified queue.
func (b *Broker) Consume(queueName string) (<-chan amqp.Delivery, error) {
	msgs, err := b.channel.Consume(
		queueName,
		"",    // consumer tag (auto-generated)
		false, // auto-ack (manual ack for reliability)
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to consume from %s: %w", queueName, err)
	}
	return msgs, nil
}

// HealthCheck verifies the RabbitMQ connection is still alive.
func (b *Broker) HealthCheck() error {
	if b.conn.IsClosed() {
		return fmt.Errorf("rabbitmq connection is closed")
	}
	return nil
}

// Close cleanly shuts down the channel and connection.
func (b *Broker) Close() {
	if b.channel != nil {
		b.channel.Close()
	}
	if b.conn != nil {
		b.conn.Close()
	}
}
