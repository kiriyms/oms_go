package main

import "context"

type OrderService interface {
	CreateOrder(context.Context) error
}

type service struct {
	store OrderStore
}

func NewOrderService(store OrderStore) OrderService {
	return &service{store: store}
}

func (s *service) CreateOrder(ctx context.Context) error {
	return s.store.Create(ctx)
}
