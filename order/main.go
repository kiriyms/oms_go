package main

import (
	"log"
	"net"

	common "github.com/kiriyms/oms_go-common"
	pb "github.com/kiriyms/oms_go-common/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	grpcAddr         = common.GetEnv("GRPC_ADDR", "localhost:50051")
	dbPath           = common.GetEnv("DB_PATH", "./db/db.db")
	stockServiceAddr = common.GetEnv("STOCK_SERVICE_ADDR", "localhost:50052")
	brokerURL        = common.GetEnv("KAFKA_BROKER_URL", "localhost:9092")
)

func main() {
	stockConn, err := grpc.NewClient(stockServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to Stock Service: %v", err)
	}
	defer stockConn.Close()
	log.Println("Dialed Stock Service at ", stockServiceAddr)

	stockC := pb.NewStockServiceClient(stockConn)

	grpcServer := grpc.NewServer()
	l, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	defer l.Close()

	store, err := NewStore(dbPath)
	if err != nil {
		log.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	producer := NewProducer(brokerURL)
	defer producer.Close()

	service := NewOrderService(store, stockC)
	NewHandler(grpcServer, service, stockC, producer)

	log.Println("gRPC server listening on", grpcAddr)

	if err := grpcServer.Serve(l); err != nil {
		log.Fatal(err.Error())
	}
}
