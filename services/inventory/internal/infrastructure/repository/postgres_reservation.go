package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/dandirahmadani19/distributed-saga-orchestrator/services/inventory/internal/domain/entity"
	"github.com/dandirahmadani19/distributed-saga-orchestrator/services/inventory/internal/domain/repository"
	"github.com/google/uuid"
)

type postgresReservationRepository struct {
	db *sql.DB
}

func NewPostgresReservationRepository(db *sql.DB) repository.ReservationRepository {
	return &postgresReservationRepository{db: db}
}
func (r *postgresReservationRepository) Create(ctx context.Context, reservation *entity.Reservation, idempotencyKey string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Generate ID (single place — no double UUID problem!)
	reservation.ID = uuid.New().String()

	// Insert reservation
	query := `
		INSERT INTO reservations (id, order_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err = tx.ExecContext(ctx, query,
		reservation.ID, reservation.OrderID, reservation.Status,
		reservation.CreatedAt, reservation.UpdatedAt,
	)
	if err != nil {
		return err
	}

	// Insert reservation items
	for _, item := range reservation.Items {
		itemQuery := `
			INSERT INTO reservation_items (reservation_id, product_id, quantity)
			VALUES ($1, $2, $3)
		`
		_, err = tx.ExecContext(ctx, itemQuery,
			reservation.ID, item.ProductID, item.Quantity,
		)
		if err != nil {
			return err
		}
	}

	// Store idempotency key
	response, _ := json.Marshal(reservation)
	idempQuery := `
		INSERT INTO reservation_idempotency (key, reservation_id, operation, response, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err = tx.ExecContext(ctx, idempQuery,
		idempotencyKey, reservation.ID, "RESERVE", response,
		time.Now(), time.Now().Add(24*time.Hour),
	)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (r *postgresReservationRepository) CheckIdempotency(ctx context.Context, key string) (*entity.Reservation, error) {
	query := `
		SELECT 
			reservations.id,
			reservations.order_id,
			reservations.status,
			reservations.created_at,
			reservations.updated_at
		FROM reservation_idempotency
		JOIN reservations ON reservation_idempotency.reservation_id = reservations.id
		WHERE reservation_idempotency.key = $1 AND reservation_idempotency.expires_at > NOW()
	`
	var reservation entity.Reservation
	var status string
	err := r.db.QueryRowContext(ctx, query, key).Scan(
		&reservation.ID, &reservation.OrderID, &status,
		&reservation.CreatedAt, &reservation.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	reservation.Status = entity.ReservationStatus(status)
	return &reservation, nil
}
