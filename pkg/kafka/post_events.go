package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	kafkago "github.com/segmentio/kafka-go"

	"SmartFeed/internal/domain"
)

type PostEventProducer struct {
	writer *kafkago.Writer
	topic  string
}

func NewPostEventProducer(brokers []string, topic string) *PostEventProducer {
	return &PostEventProducer{
		writer: &kafkago.Writer{
			Addr:         kafkago.TCP(brokers...),
			Topic:        topic,
			Balancer:     &kafkago.LeastBytes{},
			RequiredAcks: kafkago.RequireOne,
		},
		topic: topic,
	}
}

func (p *PostEventProducer) PublishPostCreated(ctx context.Context, event domain.PostCreatedEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("kafka.post_event.marshal: %w", err)
	}

	msg := kafkago.Message{
		Key:   []byte(fmt.Sprintf("post-%d", event.PostID)),
		Value: payload,
	}
	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("kafka.post_event.publish: %w", err)
	}
	return nil
}

func (p *PostEventProducer) Close() error {
	return p.writer.Close()
}

func NewPostEventReader(brokers []string, topic, groupID string) *kafkago.Reader {
	return kafkago.NewReader(kafkago.ReaderConfig{
		Brokers:  brokers,
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 1,
		MaxBytes: 10e6,
	})
}
