package main

import (
	"context"
	"time"

	pb "github.com/kiriyms/oms_go-common/api"
)

type KitchenService interface {
	AcceptOrder(context.Context, *pb.Order) error
	ProcessOrder(context.Context, *pb.Order) error
	FinishOrder(context.Context, string) error
}

type service struct {
	store Store
}

func NewService(store Store) *service {
	return &service{store: store}
}

func (s *service) AcceptOrder(ctx context.Context, o *pb.Order) error {
	s.store.AcceptOrder(ctx, o)
	return nil
}

func (s *service) ProcessOrder(ctx context.Context, o *pb.Order) error {
	time.Sleep(30 * time.Second)
	return nil
}

func (s *service) FinishOrder(ctx context.Context, orderId string) error {
	s.store.FinishOrder(ctx, orderId)
	// Send to Kafka
	return nil
}
