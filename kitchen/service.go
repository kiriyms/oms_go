package main

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "github.com/kiriyms/oms_go-common/api"
)

type KitchenService interface {
	AcceptOrder(context.Context, *pb.Order) error
	ProcessOrder(context.Context, *pb.Order) error
	FinishOrder(context.Context, string) error
}

type service struct {
	store    Store
	producer *Producer
}

func NewService(store Store, producer *Producer) *service {
	return &service{store: store, producer: producer}
}

func (s *service) AcceptOrder(ctx context.Context, o *pb.Order) error {
	err := s.store.AcceptOrder(ctx, o)
	if err != nil {
		log.Printf("Failed to accept order: %v", err)
		return err
	}
	log.Printf("Accepted order: %v", o)
	return nil
}

func (s *service) ProcessOrder(ctx context.Context, o *pb.Order) error {
	time.Sleep(10 * time.Second)
	log.Printf("Processed order: %v", o)
	return nil
}

func (s *service) FinishOrder(ctx context.Context, orderId string) error {
	o, err := s.store.GetOrder(ctx, orderId)
	if err != nil {
		return err
	}
	if o == nil {
		return fmt.Errorf("order not found: %s", orderId)
	}
	err = s.store.FinishOrder(ctx, orderId)
	if err != nil {
		return err
	}
	s.producer.PublishOrderFinished(ctx, o)
	log.Printf("Finished order: %s", orderId)
	return nil
}
