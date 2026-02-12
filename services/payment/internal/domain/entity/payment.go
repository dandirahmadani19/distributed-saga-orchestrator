package entity

import (
	"time"

	"github.com/google/uuid"
)

type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "PENDING"
	PaymentStatusProcessed PaymentStatus = "PROCESSED"
	PaymentStatusFailed    PaymentStatus = "FAILED"
	PaymentStatusRefunded  PaymentStatus = "REFUNDED"
)

type Payment struct {
	ID         string
	OrderID    string
	CustomerID string
	Amount     float64
	Status     PaymentStatus
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func NewPayment(orderID, customerID string, amount float64) *Payment {
	now := time.Now()
	return &Payment{
		ID:         uuid.New().String(),
		OrderID:    orderID,
		CustomerID: customerID,
		Amount:     amount,
		Status:     PaymentStatusPending,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

func (p *Payment) Refund() {
	p.Status = PaymentStatusRefunded
	p.UpdatedAt = time.Now()
}

func (p *Payment) Process() {
	p.Status = PaymentStatusProcessed
	p.UpdatedAt = time.Now()
}

func (p *Payment) Fail() {
	p.Status = PaymentStatusFailed
	p.UpdatedAt = time.Now()
}
