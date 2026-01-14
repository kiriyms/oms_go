package main

import (
	"context"
	"log"

	"github.com/google/uuid"
	pb "github.com/kiriyms/oms_go-common/api"
	"google.golang.org/grpc"
)

type Handler struct {
	pb.UnimplementedOrderServiceServer
	service OrderService
}

func NewHandler(s *grpc.Server, service OrderService) *Handler {
	h := &Handler{
		service: service,
	}
	pb.RegisterOrderServiceServer(s, h)
	return h
}

func (h *Handler) CreateOrder(ctx context.Context, p *pb.CreateOrderRequest) (*pb.Order, error) {
	log.Printf("New order received: %v", p)
	if err := h.service.ValidateOrder(ctx, p); err != nil {
		return nil, err
	}
	o := &pb.Order{
		ID:         uuid.New().String(),
		CustomerID: p.CustomerID,
		Status:     "PENDING",
		Items:      []*pb.Item{},
	}
	return o, nil
}
