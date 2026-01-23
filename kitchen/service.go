package main

import pb "github.com/kiriyms/oms_go-common/api"

type KitchenService interface {
	AcceptOrder(*pb.Order) error
	ProcessOrder(*pb.Order) error
	FinishOrder(string) error
}

type service struct {
	store Store
}

func NewService(store Store) *service {
	return &service{store: store}
}

func (s *service) AcceptOrder(o *pb.Order) error {
	// Add to DB queue
	return nil
}

func (s *service) ProcessOrder(o *pb.Order) error {
	// Sleep, simulate processing
	return nil
}

func (s *service) FinishOrder(orderId string) error {
	// Remove from DB queue
	// Send to Kafka
	return nil
}
