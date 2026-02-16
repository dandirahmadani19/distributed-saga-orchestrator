package entity

import (
	"time"

	pErrors "github.com/dandirahmadani19/distributed-saga-orchestrator/platform/errors"
	"github.com/google/uuid"
)

type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "PENDING"
	PaymentStatusProcessed PaymentStatus = "PROCESSING"
	PaymentStatusFailed    PaymentStatus = "FAILED"
	PaymentStatusRefunded  PaymentStatus = "REFUNDED"
	PaymentStatusCompleted PaymentStatus = "COMPLETED"
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

func NewPayment(orderID, customerID string, amount float64) (*Payment, error) {
	if orderID == "" {
		return nil, pErrors.E(pErrors.Invalid, "order id is required", nil)
	}

	if customerID == "" {
		return nil, pErrors.E(pErrors.Invalid, "customer id is required", nil)
	}

	if amount <= 0 {
		return nil, pErrors.E(pErrors.Invalid, "amount must be greater than 0", nil)
	}

	now := time.Now()
	return &Payment{
		ID:         uuid.New().String(),
		OrderID:    orderID,
		CustomerID: customerID,
		Amount:     amount,
		Status:     PaymentStatusPending,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
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
