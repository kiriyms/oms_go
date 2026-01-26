package main

import (
	"context"
	"log"
	"path/filepath"

	common "github.com/kiriyms/oms_go-common"
)

var (
	dbPath    = common.GetEnv("DB_PATH", "./db/db.db")
	brokerURL = common.GetEnv("KAFKA_BROKER_URL", "localhost:9092")
)

func main() {
	abs, _ := filepath.Abs(dbPath)
	log.Println("Using DB at:", abs)
	
	store, err := NewStore(dbPath)
	if err != nil {
		log.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	producer := NewProducer(brokerURL)
	defer producer.Close()

	service := NewService(store, producer)

	consumer := NewConsumer(brokerURL, "kitchen-service", service)
	defer consumer.Close()

	ctx := context.Background()
	consumer.Start(ctx)
}
