package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	pb "github.com/kiriyms/oms_go-common/api"
	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
}

func NewConsumer(brokerURL string, groupID string) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{brokerURL},
		Topic:    "orders.created",
		GroupID:  groupID,
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})

	return &Consumer{reader: reader}
}

func (c *Consumer) Start(ctx context.Context) {
	log.Printf("Starting consumer...")
	for {
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("error reading message: %v (retrying in 5s)", err)
			time.Sleep(5 * time.Second)
			continue
		}

		var event pb.Order
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("failed to unmarshal event: %v", err)
			continue
		}

		log.Printf("Received order %s", event.ID)
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
