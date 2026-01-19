package main

import (
	"context"

	pb "github.com/kiriyms/oms_go-common/api"
)

type StockService interface {
	AddStockItem(ctx context.Context, item *pb.AddStockItemRequest) error
	BookStockItems(ctx context.Context, item *pb.BookItemsRequest) error
	ReleaseBookItems(ctx context.Context, item *pb.ReleaseBookedItemsRequest) error
	RemoveStockItem(ctx context.Context, itemID string) error
	VerifyStock(ctx context.Context, req *pb.VerifyStockRequest) *pb.VerifyStockResponse
	GetStockItem(ctx context.Context, req *pb.GetStockItemRequest) (*pb.StockItem, error)
}

type service struct {
	store StockStore
}

func NewStockService(store StockStore) *service {
	return &service{store: store}
}
