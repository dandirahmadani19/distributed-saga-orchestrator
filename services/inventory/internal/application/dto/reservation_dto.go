package dto

import "time"

// ReserveInventoryRequest is the input for reserving inventory
type ReserveInventoryRequest struct {
	IdempotencyKey string
	OrderID        string
	Items          []ReserveItemRequest
}

// ReserveItemRequest represents one item to reserve
type ReserveItemRequest struct {
	ProductID string
	Quantity  int
}

// ReleaseInventoryRequest is the input for releasing inventory
type ReleaseInventoryRequest struct {
	IdempotencyKey string
	OrderID        string
}

// ReservationResponse is the output after reserving/releasing inventory
type ReservationResponse struct {
	ID        string
	OrderID   string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}
