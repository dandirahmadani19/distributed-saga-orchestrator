package entity

import "time"

// OrderStatus represents the state of an order
type OrderStatus string

const (
	OrderStatusCreated   OrderStatus = "CREATED"
	OrderStatusConfirmed OrderStatus = "CONFIRMED"
	OrderStatusCancelled OrderStatus = "CANCELLED"
)

// Order represents an order aggregate root
type Order struct {
	ID          string
	CustomerID  string
	Items       []OrderItem
	TotalAmount float64
	Status      OrderStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// OrderItem represents one item in an order
type OrderItem struct {
	ProductID string
	Quantity  int
	Price     float64
}

// NewOrder creates a new order (factory function)
func NewOrder(customerID string, items []OrderItem, totalAmount float64) *Order {
	now := time.Now()
	return &Order{
		CustomerID:  customerID,
		Items:       items,
		TotalAmount: totalAmount,
		Status:      OrderStatusCreated,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// Cancel marks the order as cancelled
func (o *Order) Cancel() {
	o.Status = OrderStatusCancelled
	o.UpdatedAt = time.Now()
}

// Confirm marks the order as confirmed
func (o *Order) Confirm() {
	o.Status = OrderStatusConfirmed
	o.UpdatedAt = time.Now()
}
