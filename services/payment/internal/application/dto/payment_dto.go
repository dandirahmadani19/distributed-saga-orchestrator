package dto

import "time"

// CreatePaymentRequest is the input for creating a payment
type CreatePaymentRequest struct {
	IdempotencyKey string
	CustomerID     string
	OrderID        string
	Amount         float64
}

type RefundPaymentRequest struct {
	IdempotencyKey string
	PaymentID      string
}

// PaymentResponse is the output after creating/fetching a payment
type PaymentResponse struct {
	ID         string
	OrderID    string
	CustomerID string
	Amount     float64
	Status     string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
