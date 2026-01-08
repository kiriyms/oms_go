package main

import "context"

type OrderStore interface {
	Create(context.Context) error
}

type store struct {
}

func NewStore() OrderStore {
	return &store{}
}

func (s *store) Create(ctx context.Context) error {
	return nil
}
