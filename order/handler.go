package main

import (
	"context"
	"log"

	pb "github.com/kiriyms/oms_go-common/api"
	"google.golang.org/grpc"
)

type Handler struct {
	pb.UnimplementedOrderServiceServer
}

func NewHandler(s *grpc.Server) *Handler {
	h := &Handler{}
	pb.RegisterOrderServiceServer(s, h)
	return h
}

func (h *Handler) CreateOrder(ctx context.Context, p *pb.CreateOrderRequest) (*pb.Order, error) {
	log.Printf("New order received: %v", p)
	o := &pb.Order{
		ID:         "123",
		CustomerID: "jd456",
		Status:     "completed",
		Items:      []*pb.Item{},
	}
	return o, nil
}
