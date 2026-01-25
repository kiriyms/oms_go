package main

import (
	"context"
	"log"

	common "github.com/kiriyms/oms_go-common"
)

var (
	dbPath    = common.GetEnv("DB_PATH", "./db/db.db")
	brokerURL = common.GetEnv("KAFKA_BROKER_URL", "localhost:9092")
)

func main() {
	store, err := NewStore(dbPath)
	if err != nil {
		log.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	NewService(store)

	consumer := NewConsumer(brokerURL, "kitchen-service")
	defer consumer.Close()

	ctx := context.Background()
	consumer.Start(ctx)
}
