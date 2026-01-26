package main

import (
	"context"
	"encoding/json"
	"log"

	pb "github.com/kiriyms/oms_go-common/api"
	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokerURL string) *Producer {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{brokerURL},
		Topic:    "orders.finished",
		Balancer: &kafka.LeastBytes{},
	})

	return &Producer{writer: writer}
}

func (p *Producer) PublishOrderFinished(ctx context.Context, order *pb.Order) error {
	log.Printf("Publishing order.finished event: %v", order)
	valueBytes, err := json.Marshal(order)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Key:   []byte(order.ID),
		Value: valueBytes,
	}

	err = p.writer.WriteMessages(ctx, msg)
	if err != nil {
		log.Printf("failed to write message: %v", err)
		return err
	}

	log.Printf("published order.finished event: %s", order.ID)
	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
