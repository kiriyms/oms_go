package main

import (
	"fmt"
	"net/http"

	common "github.com/kiriyms/oms_go-common"
	pb "github.com/kiriyms/oms_go-common/api"
)

type handler struct {
	client pb.OrderServiceClient
}

func NewHandler(client pb.OrderServiceClient) *handler {
	return &handler{
		client: client,
	}
}

func (h *handler) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/customers/{customerID}/order", h.HandleCreateOrder)
}

func (h *handler) HandleCreateOrder(w http.ResponseWriter, r *http.Request) {
	cID := r.PathValue("customerID")

	var items []*pb.ItemWithQuantity
	if err := common.ReadJSON(r, &items); err != nil {
		common.WriteError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request payload: %v", err))
		return
	}

	h.client.CreateOrder(r.Context(), &pb.CreateOrderRequest{
		CustomerID: cID,
		Items:      items,
	})
}
