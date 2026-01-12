package main

import (
	"log"
	"net/http"

	_ "github.com/joho/godotenv/autoload"
	common "github.com/kiriyms/oms_go-common"
	pb "github.com/kiriyms/oms_go-common/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	httpAddr         = common.GetEnv("HTTP_ADDR", ":8080")
	orderServiceAddr = common.GetEnv("ORDER_SERVICE_ADDR", "localhost:50051")
)

func main() {
	conn, err := grpc.NewClient(orderServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to Order Service: %v", err)
	}
	defer conn.Close()

	log.Println("Dialed Order Service at ", orderServiceAddr)

	c := pb.NewOrderServiceClient(conn)

	mux := http.NewServeMux()
	handler := NewHandler(c)
	handler.registerRoutes(mux)

	log.Printf("Starting Server at %s", httpAddr)

	if err := http.ListenAndServe(httpAddr, mux); err != nil {
		log.Fatal("Failed to start Gateway HTTP server")
	}
}
