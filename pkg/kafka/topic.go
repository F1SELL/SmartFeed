package kafka

import (
	"context"
	"fmt"
	"time"

	kafkago "github.com/segmentio/kafka-go"
)

func EnsureTopic(ctx context.Context, brokers []string, topic string, partitions, replicationFactor int) error {
	if len(brokers) == 0 {
		return fmt.Errorf("kafka.ensure_topic: empty brokers list")
	}
	if partitions <= 0 {
		partitions = 1
	}
	if replicationFactor <= 0 {
		replicationFactor = 1
	}

	var lastErr error
	for i := 0; i < 8; i++ {
		err := ensureTopicOnce(brokers[0], topic, partitions, replicationFactor)
		if err == nil {
			return nil
		}
		lastErr = err

		select {
		case <-ctx.Done():
			return fmt.Errorf("kafka.ensure_topic: %w", ctx.Err())
		case <-time.After(500 * time.Millisecond):
		}
	}

	return fmt.Errorf("kafka.ensure_topic: %w", lastErr)
}

func ensureTopicOnce(broker, topic string, partitions, replicationFactor int) error {
	conn, err := kafkago.Dial("tcp", broker)
	if err != nil {
		return err
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return err
	}

	controllerConn, err := kafkago.Dial("tcp", fmt.Sprintf("%s:%d", controller.Host, controller.Port))
	if err != nil {
		return err
	}
	defer controllerConn.Close()

	return controllerConn.CreateTopics(kafkago.TopicConfig{
		Topic:             topic,
		NumPartitions:     partitions,
		ReplicationFactor: replicationFactor,
	})
}
