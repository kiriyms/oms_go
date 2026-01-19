package main

import (
	"context"

	pb "github.com/kiriyms/oms_go-common/api"
)

type StockStore interface {
	AddStockItem(ctx context.Context, item *pb.StockItem) error
	BookStockItem(ctx context.Context, itemID string, quantity int32) error
	ReleaseBookItem(ctx context.Context, itemID string, quantity int32) error
	RemoveStockItem(ctx context.Context, itemID string) error
	VerifyStock(ctx context.Context, items []*pb.StockItem) *pb.VerifyStockResponse
	GetStockItem(ctx context.Context, itemID string) (*pb.StockItem, error)
	Close() error
}

type store struct{}

func NewStore(dbPath string) (*store, error) {
	return &store{}, nil
}

func (s *store) Close() error {
	return nil
}
