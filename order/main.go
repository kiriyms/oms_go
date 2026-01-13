package main

import (
	"context"
	"log"
	"net"

	common "github.com/kiriyms/oms_go-common"
	"google.golang.org/grpc"
)

var (
	grpcAddr = common.GetEnv("GRPC_ADDR", "localhost:50051")
)

func main() {
	grpcServer := grpc.NewServer()
	l, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	defer l.Close()

	store := NewStore()
	service := NewOrderService(store)
	NewHandler(grpcServer)

	service.CreateOrder(context.Background())

	log.Println("gRPC server listening on", grpcAddr)

	if err := grpcServer.Serve(l); err != nil {
		log.Fatal(err.Error())
	}
}
