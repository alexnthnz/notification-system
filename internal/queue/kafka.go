package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/alexnthnz/notification-system/internal/config"
	"github.com/segmentio/kafka-go"
)

// NotificationMessage represents a message in the notification queue
type NotificationMessage struct {
	ID        string            `json:"id"`
	UserID    string            `json:"user_id"`
	Channel   string            `json:"channel"`
	Recipient string            `json:"recipient"`
	Subject   string            `json:"subject,omitempty"`
	Body      string            `json:"body"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	Priority  int               `json:"priority"` // 1 = high, 2 = medium, 3 = low
	CreatedAt time.Time         `json:"created_at"`
}

// Producer handles publishing messages to Kafka
type Producer struct {
	writer *kafka.Writer
}

// Consumer handles consuming messages from Kafka
type Consumer struct {
	reader *kafka.Reader
}

// NewProducer creates a new Kafka producer
func NewProducer(cfg config.KafkaConfig) *Producer {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(cfg.Brokers...),
		Topic:    cfg.Topic,
		Balancer: &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
		BatchSize:    100,
		Async:        false, // Synchronous for reliability
	}

	return &Producer{writer: writer}
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(cfg config.KafkaConfig, groupID string) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     cfg.Brokers,
		Topic:       cfg.Topic,
		GroupID:     groupID,
		MinBytes:    10e3, // 10KB
		MaxBytes:    10e6, // 10MB
		MaxWait:     1 * time.Second,
		StartOffset: kafka.LastOffset,
	})

	return &Consumer{reader: reader}
}

// PublishNotification publishes a notification message to Kafka
func (p *Producer) PublishNotification(ctx context.Context, msg NotificationMessage) error {
	// Marshal the message to JSON
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal notification message: %w", err)
	}

	// Create Kafka message
	kafkaMsg := kafka.Message{
		Key:   []byte(msg.ID),
		Value: data,
		Headers: []kafka.Header{
			{Key: "channel", Value: []byte(msg.Channel)},
			{Key: "priority", Value: []byte(fmt.Sprintf("%d", msg.Priority))},
		},
		Time: time.Now(),
	}

	// Write message to Kafka
	if err := p.writer.WriteMessages(ctx, kafkaMsg); err != nil {
		return fmt.Errorf("failed to write message to Kafka: %w", err)
	}

	log.Printf("Published notification %s to Kafka topic", msg.ID)
	return nil
}

// ConsumeNotifications consumes notification messages from Kafka
func (c *Consumer) ConsumeNotifications(ctx context.Context, handler func(NotificationMessage) error) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Read message from Kafka
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				log.Printf("Error reading message from Kafka: %v", err)
				continue
			}

			// Unmarshal the notification message
			var notification NotificationMessage
			if err := json.Unmarshal(msg.Value, &notification); err != nil {
				log.Printf("Error unmarshaling notification message: %v", err)
				continue
			}

			// Process the message
			if err := handler(notification); err != nil {
				log.Printf("Error processing notification %s: %v", notification.ID, err)
				// In a production system, you might want to implement a dead letter queue here
				continue
			}

			log.Printf("Successfully processed notification %s", notification.ID)
		}
	}
}

// Close closes the producer
func (p *Producer) Close() error {
	return p.writer.Close()
}

// Close closes the consumer
func (c *Consumer) Close() error {
	return c.reader.Close()
}