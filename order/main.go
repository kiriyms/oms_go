package main

import "context"

func main() {
	store := NewStore()
	service := NewOrderService(store)

	service.CreateOrder(context.Background())
}
