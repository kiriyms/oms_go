package main

import (
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
