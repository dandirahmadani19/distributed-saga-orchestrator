package dto

import "time"

// CreateOrderRequest is the input for creating an order
type CreateOrderRequest struct {
	IdempotencyKey string
	CustomerID     string
	Items          []OrderItemDTO
	TotalAmount    float64
}

// OrderItemDTO represents an order item in DTOs
type OrderItemDTO struct {
	ProductID string
	Quantity  int
	Price     float64
}

// OrderResponse is the output after creating/fetching an order
type OrderResponse struct {
	ID          string
	CustomerID  string
	Status      string
	TotalAmount float64
	CreatedAt   time.Time
}
