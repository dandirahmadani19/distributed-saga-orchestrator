package entity

import "time"

// ReservationStatus represents the state of a reservation
type ReservationStatus string

const (
	ReservationStatusReserved  ReservationStatus = "RESERVED"
	ReservationStatusReleased  ReservationStatus = "RELEASED"
	ReservationStatusConfirmed ReservationStatus = "CONFIRMED"
)

// Reservation represents an inventory reservation
// Analogy: This is like a "Reserved" sign on a restaurant table.
// It says: "These items belong to this order, don't give them to anyone else."
type Reservation struct {
	ID        string
	OrderID   string
	Items     []ReservationItem
	Status    ReservationStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ReservationItem represents one item being reserved
// Analogy: Each line item on the reservation slip — "2x Chicken, 1x Rice"
type ReservationItem struct {
	ProductID string
	Quantity  int
}

// NewReservation creates a new reservation (factory function)
// Note: ID is NOT set here — the repository will set it (lesson from payment service!)
func NewReservation(orderID string, items []ReservationItem) *Reservation {
	now := time.Now()
	return &Reservation{
		OrderID:   orderID,
		Items:     items,
		Status:    ReservationStatusReserved,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Release marks the reservation as released (compensation)
// Analogy: Remove the "Reserved" sign from the table. Other customers can now sit there.
func (r *Reservation) Release() {
	r.Status = ReservationStatusReleased
	r.UpdatedAt = time.Now()
}

// Confirm marks the reservation as confirmed (saga completed successfully)
func (r *Reservation) Confirm() {
	r.Status = ReservationStatusConfirmed
	r.UpdatedAt = time.Now()
}
