package main

import (
	"context"

	pb "github.com/kiriyms/oms_go-common/api"
	"google.golang.org/grpc"
)

type Handler struct {
	pb.UnimplementedStockServiceServer
	service StockService
}

func NewHandler(s *grpc.Server, service StockService) *Handler {
	h := &Handler{
		service: service,
	}
	pb.RegisterStockServiceServer(s, h)
	return h
}

func (h *Handler) AddStockItem(ctx context.Context, req *pb.AddStockItemRequest) (*pb.AddStockItemResponse, error) {
	item, err := h.service.AddStockItem(ctx, req)
	if err != nil {
		return nil, err
	}
	return &pb.AddStockItemResponse{Item: item}, nil
}

func (h *Handler) BookItems(ctx context.Context, req *pb.BookItemsRequest) (*pb.BookItemsResponse, error) {
	items, err := h.service.BookStockItems(ctx, req)
	if err != nil {
		return nil, err
	}
	return &pb.BookItemsResponse{Bookings: items}, nil
}

func (h *Handler) ReleaseBookedItems(ctx context.Context, req *pb.ReleaseBookedItemsRequest) (*pb.ReleaseBookedItemsResponse, error) {
	_, err := h.service.ReleaseBookItems(ctx, req)
	if err != nil {
		return nil, err
	}
	return &pb.ReleaseBookedItemsResponse{Success: true}, nil
}

func (h *Handler) RemoveStockItem(ctx context.Context, req *pb.RemoveStockItemRequest) (*pb.RemoveStockItemResponse, error) {
	item, err := h.service.RemoveStockItem(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	return &pb.RemoveStockItemResponse{Item: item}, nil
}

func (h *Handler) VerifyStock(ctx context.Context, req *pb.VerifyStockRequest) (*pb.VerifyStockResponse, error) {
	return h.service.VerifyStock(ctx, req)
}

func (h *Handler) GetStockItem(ctx context.Context, req *pb.GetStockItemRequest) (*pb.GetStockItemResponse, error) {
	item, err := h.service.GetStockItem(ctx, req)
	if err != nil {
		return nil, err
	}
	return &pb.GetStockItemResponse{Item: item}, nil
}

func (h *Handler) FinalizeBooking(ctx context.Context, req *pb.FinalizeBookingRequest) (*pb.FinalizeBookingResponse, error) {
	err := h.service.FinalizeBooking(ctx, req.OrderID)
	if err != nil {
		return nil, err
	}
	return &pb.FinalizeBookingResponse{Success: true}, nil
}
