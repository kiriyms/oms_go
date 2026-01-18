package main

import (
	"context"
	"log"

	common "github.com/kiriyms/oms_go-common"
	pb "github.com/kiriyms/oms_go-common/api"
)

type OrderService interface {
	CreateOrder(context.Context, *pb.Order) error
	ValidateOrder(context.Context, *pb.CreateOrderRequest) error
	GetOrder(context.Context, string) (*pb.Order, error)
	GetUserOrders(context.Context, string) ([]*pb.Order, error)
	PatchOrderStatus(context.Context, string, pb.OrderStatus) (*pb.Order, error)
}

type service struct {
	store OrderStore
}

func NewOrderService(store OrderStore) *service {
	return &service{store: store}
}

func (s *service) CreateOrder(ctx context.Context, o *pb.Order) error {
	return s.store.Create(ctx, o)
}

func (s *service) ValidateOrder(ctx context.Context, p *pb.CreateOrderRequest) error {
	if len(p.Items) == 0 {
		return common.ErrNoItems
	}

	merged := mergeItemsQuantities(p.Items)
	log.Printf("Merged items: %v", merged)

	// validate with store service

	return nil
}

func (s *service) GetOrder(ctx context.Context, id string) (*pb.Order, error) {
	return s.store.GetOrder(ctx, id)
}

func (s *service) GetUserOrders(ctx context.Context, customerID string) ([]*pb.Order, error) {
	return s.store.GetUserOrders(ctx, customerID)
}

func (s *service) PatchOrderStatus(ctx context.Context, orderID string, status pb.OrderStatus) (*pb.Order, error) {
	return s.store.PatchOrderStatus(ctx, orderID, status)
}

func mergeItemsQuantities(items []*pb.ItemWithQuantity) []*pb.ItemWithQuantity {
	merged := make([]*pb.ItemWithQuantity, 0)
	itemMap := make(map[string]int32)

	for _, item := range items {
		itemMap[item.ID] += item.Quantity
	}

	for id, quantity := range itemMap {
		merged = append(merged, &pb.ItemWithQuantity{
			ID:       id,
			Quantity: quantity,
		})
	}

	return merged
}
