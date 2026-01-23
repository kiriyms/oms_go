package main

import (
	"log"

	common "github.com/kiriyms/oms_go-common"
)

var (
	dbPath = common.GetEnv("DB_PATH", "./db/db.db")
)

func main() {
	store, err := NewStore(dbPath)
	if err != nil {
		log.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	NewService(store)
}
