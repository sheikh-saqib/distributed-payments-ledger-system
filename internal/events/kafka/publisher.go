package kafka

import (
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"
)

type Publisher struct {
	writer *kafka.Writer
}

func NewPublisher(brokers []string) *Publisher {
	return &Publisher{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Topic:    "transaction_completed",
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (p *Publisher) Publish(topic string, event any) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return p.writer.WriteMessages(
		context.Background(),
		kafka.Message{
			Value: data,
		},
	)
}
