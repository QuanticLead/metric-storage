package main

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

var (
	kafkaProducer      *kafka.Producer
	kafkaFallbackTopic string
)

// InitKafka initializes the global Kafka producer if brokers are configured.
func InitKafka(cfg *Config) error {
	if cfg.KafkaBrokers == "" {
		slog.Info("Kafka brokers not configured. Kafka fallback is disabled.")
		return nil
	}

	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": cfg.KafkaBrokers,
	})
	if err != nil {
		return err
	}

	kafkaProducer = producer
	kafkaFallbackTopic = cfg.KafkaFallbackTopic
	slog.Info("Kafka fallback producer initialized", slog.String("brokers", cfg.KafkaBrokers), slog.String("topic", kafkaFallbackTopic))

	// Start a goroutine to read delivery reports and log any failures.
	go func() {
		for e := range producer.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					slog.Error("Kafka delivery failed", "error", ev.TopicPartition.Error)
				} else {
					slog.Debug("Kafka message delivered", slog.String("topic", *ev.TopicPartition.Topic), slog.Int("partition", int(ev.TopicPartition.Partition)))
				}
			}
		}
	}()

	return nil
}

// CloseKafka flushes and closes the global Kafka producer.
func CloseKafka() {
	if kafkaProducer != nil {
		slog.Info("Flushing and closing Kafka producer...")
		kafkaProducer.Flush(5000)
		kafkaProducer.Close()
		kafkaProducer = nil
	}
}

// FlushKafka blocks until all outstanding messages are delivered.
func FlushKafka(timeoutMs int) int {
	if kafkaProducer != nil {
		return kafkaProducer.Flush(timeoutMs)
	}
	return 0
}

// PublishFallbackMetrics serializes metrics and publishes them to the Kafka fallback topic.
func PublishFallbackMetrics(ctx context.Context, metricType string, rawMetrics interface{}) error {
	if kafkaProducer == nil {
		slog.WarnContext(ctx, "Kafka producer not initialized. Dropping metric fallback batch.", slog.String("type", metricType))
		return nil
	}

	payload := map[string]interface{}{
		"type":    metricType,
		"metrics": rawMetrics,
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to marshal metrics payload for Kafka fallback", "error", err)
		return err
	}

	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &kafkaFallbackTopic,
			Partition: kafka.PartitionAny,
		},
		Value: jsonBytes,
	}

	err = kafkaProducer.Produce(msg, nil)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to produce fallback metrics to Kafka", "error", err)
		return err
	}

	return nil
}
