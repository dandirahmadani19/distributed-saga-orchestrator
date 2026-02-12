package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/dandirahmadani19/distributed-saga-orchestrator/services/payment/internal/domain/entity"
	"github.com/dandirahmadani19/distributed-saga-orchestrator/services/payment/internal/domain/repository"
	"github.com/google/uuid"
)

type postgresPaymentRepository struct {
	db *sql.DB
}

func NewPostgresPaymentRepository(db *sql.DB) repository.PaymentRepository {
	return &postgresPaymentRepository{db: db}
}

func (r *postgresPaymentRepository) Create(ctx context.Context, payment *entity.Payment, idempotencyKey string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Generate ID
	payment.ID = uuid.New().String()

	// Insert payment
	query := `
		INSERT INTO payments (id, customer_id, order_id, status, amount, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err = tx.ExecContext(ctx, query,
		payment.ID, payment.CustomerID, payment.OrderID, payment.Status, payment.Amount,
		payment.CreatedAt, payment.UpdatedAt,
	)
	if err != nil {
		return err
	}

	// Store idempotency key
	response, _ := json.Marshal(payment)
	idempQuery := `
		INSERT INTO payment_idempotency (key, payment_id, operation, response, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err = tx.ExecContext(ctx, idempQuery,
		idempotencyKey, payment.ID, "CREATE", response,
		time.Now(), time.Now().Add(24*time.Hour),
	)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (r *postgresPaymentRepository) CheckIdempotency(ctx context.Context, key string) (*entity.Payment, error) {
	query := `
		SELECT 
			payments.id,
			payments.customer_id,
			payments.order_id,
			payments.status,
			payments.amount,
			payments.created_at,
			payments.updated_at
		FROM payment_idempotency
		JOIN payments ON payment_idempotency.payment_id = payments.id
		WHERE payment_idempotency.key = $1 AND payment_idempotency.expires_at > NOW()
	`

	var payment entity.Payment
	var status string

	err := r.db.QueryRowContext(ctx, query, key).Scan(
		&payment.ID, &payment.CustomerID, &payment.OrderID, &status, &payment.Amount,
		&payment.CreatedAt, &payment.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	payment.Status = entity.PaymentStatus(status)
	return &payment, nil
}
