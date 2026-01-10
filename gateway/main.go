package main

import (
	"log"
	"net/http"

	_ "github.com/joho/godotenv/autoload"
	common "github.com/kiriyms/oms_go-common"
)

var (
	httpAddr = common.GetEnv("HTTP_ADDR", ":8080")
)

func main() {
	mux := http.NewServeMux()
	handler := NewHandler()
	handler.registerRoutes(mux)

	log.Printf("Starting Server at %s", httpAddr)

	if err := http.ListenAndServe(httpAddr, mux); err != nil {
		log.Fatal("Failed to start Gateway HTTP server")
	}
}
