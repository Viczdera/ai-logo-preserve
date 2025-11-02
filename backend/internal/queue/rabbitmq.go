package queue

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Viczdera/ai-logo-preserve/backend/internal/models"
	"github.com/Viczdera/ai-logo-preserve/backend/internal/utils"

	"github.com/redis/go-redis/v9"
	"github.com/streadway/amqp"
)

type Client interface {
	PublishJob(job *models.Job) error
	ConsumeJobs(handler func(*models.Job) error) error
	Close() error
}

type RabbitMQClient struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	exchange  string
	queueName string
}

func NewRabbitMQClient(cfg utils.RabbitMQConfig) (Client, error) {
	// Connect to RabbitMQ
	conn, err := amqp.Dial(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	err = channel.ExchangeDeclare(
		cfg.Exchange, // name
		"direct",     // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Declare queue
	queue, err := channel.QueueDeclare(
		cfg.Queue, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	// Bind queue to exchange
	err = channel.QueueBind(
		queue.Name,   // queue name
		"detection",  // routing key
		cfg.Exchange, // exchange
		false,
		nil,
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to bind queue: %w", err)
	}

	return &RabbitMQClient{
		conn:      conn,
		channel:   channel,
		exchange:  cfg.Exchange,
		queueName: queue.Name,
	}, nil
}

func (c *RabbitMQClient) PublishJob(job *models.Job) error {
	// Serialize job to JSON
	body, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}
	fmt.Printf("Published job: %s", string(body))

	// Publish message
	err = c.channel.Publish(
		c.exchange,  // exchange
		"detection", // routing key
		false,       // mandatory
		false,       // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent, // Make message persistent
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

func (c *RabbitMQClient) ConsumeJobs(handler func(*models.Job) error) error {
	// Set QoS to process one message at a time
	err := c.channel.Qos(1, 0, false)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	// Start consuming messages
	msgs, err := c.channel.Consume(
		c.queueName, // queue
		"",          // consumer
		false,       // auto-ack
		false,       // exclusive
		false,       // no-local
		false,       // no-wait
		nil,         // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	// Process messages
	go func() {
		for msg := range msgs {
			var job models.Job
			if err := json.Unmarshal(msg.Body, &job); err != nil {
				fmt.Printf("Failed to unmarshal job: %v\n", err)
				msg.Nack(false, false) // Reject message
				continue
			}

			// Process job
			if err := handler(&job); err != nil {
				fmt.Printf("Failed to process job %s: %v\n", job.ID, err)
				msg.Nack(false, true) // Reject and requeue
				continue
			}

			// Acknowledge message
			msg.Ack(false)
		}
	}()

	return nil
}

func (c *RabbitMQClient) Close() error {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Redis client for rate limiting
func NewRedisClient(cfg utils.RedisConfig) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
}
