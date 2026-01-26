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
	reader  *kafka.Reader
	service KitchenService
}

func NewConsumer(brokerURL string, groupID string, service KitchenService) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{brokerURL},
		Topic:    "orders.created",
		GroupID:  groupID,
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})

	return &Consumer{reader: reader, service: service}
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

		go func() {
			log.Printf("Received order %s", event.ID)
			if err := c.service.AcceptOrder(ctx, &event); err != nil {
				log.Printf("failed to accept order: %v", err)
			}

			if err := c.service.ProcessOrder(ctx, &event); err != nil {
				log.Printf("failed to process order: %v", err)
			}

			if err := c.service.FinishOrder(ctx, event.ID); err != nil {
				log.Printf("failed to finish order: %v", err)
			}
		}()
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
