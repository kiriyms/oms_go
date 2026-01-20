package main

import (
	"context"

	pb "github.com/kiriyms/oms_go-common/api"
)

type StockService interface {
	AddStockItem(ctx context.Context, item *pb.AddStockItemRequest) (*pb.StockItem, error)
	BookStockItems(ctx context.Context, item *pb.BookItemsRequest) ([]*pb.ItemWithQuantity, error)
	ReleaseBookItems(ctx context.Context, item *pb.ReleaseBookedItemsRequest) ([]*pb.ItemWithQuantity, error)
	RemoveStockItem(ctx context.Context, itemID string) (*pb.StockItem, error)
	VerifyStock(ctx context.Context, req *pb.VerifyStockRequest) (*pb.VerifyStockResponse, error)
	GetStockItem(ctx context.Context, req *pb.GetStockItemRequest) (*pb.StockItem, error)
	FinalizeBooking(ctx context.Context, orderID string) error
}

type service struct {
	store StockStore
}

func NewStockService(store StockStore) *service {
	return &service{store: store}
}

func (s *service) AddStockItem(ctx context.Context, req *pb.AddStockItemRequest) (*pb.StockItem, error) {
	stockItem := &pb.StockItem{
		ID:          req.ID,
		Quantity:    req.Quantity,
		Name:        req.Name,
		PriceID:     req.PriceID,
		Description: req.Description,
		ImgPath:     req.ImgPath,
	}

	return s.store.AddStockItem(ctx, stockItem)
}

func (s *service) BookStockItems(ctx context.Context, req *pb.BookItemsRequest) ([]*pb.ItemWithQuantity, error) {
	var bookedItems []*pb.ItemWithQuantity
	for _, item := range req.Items {
		bItem, err := s.store.BookStockItem(ctx, item.ID, item.Quantity)
		if err != nil {
			return nil, err
		}
		bookedItems = append(bookedItems, bItem)
	}
	return bookedItems, nil
}

func (s *service) ReleaseBookItems(ctx context.Context, req *pb.ReleaseBookedItemsRequest) ([]*pb.ItemWithQuantity, error) {
	var releasedItems []*pb.ItemWithQuantity
	for _, item := range req.Items {
		rItems, err := s.store.ReleaseBookItem(ctx, item.ID, item.Quantity)
		if err != nil {
			return nil, err
		}
		releasedItems = append(releasedItems, rItems)
	}
	return releasedItems, nil
}

func (s *service) RemoveStockItem(ctx context.Context, itemID string) (*pb.StockItem, error) {
	return s.store.RemoveStockItem(ctx, itemID)
}

func (s *service) VerifyStock(ctx context.Context, req *pb.VerifyStockRequest) (*pb.VerifyStockResponse, error) {
	return s.store.VerifyStock(ctx, req.Items), nil
}

func (s *service) GetStockItem(ctx context.Context, req *pb.GetStockItemRequest) (*pb.StockItem, error) {
	return s.store.GetStockItem(ctx, req.ID)
}

func (s *service) FinalizeBooking(ctx context.Context, orderID string) error {
	return s.store.FinalizeBooking(ctx, orderID)
}
