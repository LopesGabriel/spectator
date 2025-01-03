package main

import (
	"log/slog"
	"os"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func main() {
	if len(os.Args) < 2 {
		slog.Error("missing command")
		os.Exit(1)
	}

	kafkaHost := getEnvOrDefault("KAFKA_HOST", "localhost:9092")

	p, err := kafka.NewProducer(&kafka.ConfigMap{
		// User-specific properties that you must set
		"bootstrap.servers": kafkaHost,

		// Fixed properties
		"acks": "all",
	})

	if err != nil {
		slog.Error("failed to create producer", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Go-routine to handle message delivery reports and
	// possibly other event types (errors, stats, etc)
	go func() {
		for e := range p.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					slog.Warn("failed to deliver message", "topic-partition", ev.TopicPartition)
				} else {
					slog.Info(
						"Produced event to topic",
						slog.String("topic", *ev.TopicPartition.Topic),
						slog.String("key", string(ev.Key)),
						slog.String("value", string(ev.Value)),
					)
				}
			}
		}
	}()

	topic := "spectator-topic"
	command := os.Args[1]

	p.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          []byte(command),
	}, nil)

	// Wait for all messages to be delivered
	p.Flush(5 * 1000)
	p.Close()
}

func getEnvOrDefault(env, defaultValue string) string {
	value, isPresent := os.LookupEnv(env)

	if !isPresent {
		return defaultValue
	}

	return value
}
