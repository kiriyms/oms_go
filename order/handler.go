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
		Items:      h.mapItemWithQuantityToItem(p.Items),
	}

	err := h.service.CreateOrder(ctx, o)
	if err != nil {
		return nil, err
	}
	return o, nil
}

func (h *Handler) mapItemWithQuantityToItem(iwq []*pb.ItemWithQuantity) []*pb.Item {
	items := make([]*pb.Item, 0)
	for _, item := range iwq {
		items = append(items, &pb.Item{
			ID:       item.ID,
			Quantity: item.Quantity,
		})
	}
	return items
}
