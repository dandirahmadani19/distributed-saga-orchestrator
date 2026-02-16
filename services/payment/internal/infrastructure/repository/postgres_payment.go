package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/dandirahmadani19/distributed-saga-orchestrator/services/payment/internal/domain/entity"
	"github.com/dandirahmadani19/distributed-saga-orchestrator/services/payment/internal/domain/repository"

	pErrors "github.com/dandirahmadani19/distributed-saga-orchestrator/platform/errors"
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
		return pErrors.E(pErrors.Internal, "failed to begin transaction", err)
	}
	defer tx.Rollback()

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
		return pErrors.E(pErrors.Internal, "failed to insert payment", err)
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
		return pErrors.E(pErrors.Internal, "failed to store idempotency key", err)
	}

	if err := tx.Commit(); err != nil {
		return pErrors.E(pErrors.Internal, "failed to commit transaction", err)
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
		return nil, pErrors.E(pErrors.Internal, "failed to check idempotency", err)
	}

	payment.Status = entity.PaymentStatus(status)
	return &payment, nil
}

func (r *postgresPaymentRepository) GetByID(ctx context.Context, id string) (*entity.Payment, error) {
	query := `
		SELECT 
			id,
			customer_id,
			order_id,
			status,
			amount,
			created_at,
			updated_at
		FROM payments
		WHERE id = $1
	`

	var payment entity.Payment
	var status string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&payment.ID, &payment.CustomerID, &payment.OrderID, &status, &payment.Amount,
		&payment.CreatedAt, &payment.UpdatedAt,
	)

	if err != nil {
		return nil, pErrors.E(pErrors.Internal, "failed to get payment by id", err)
	}

	payment.Status = entity.PaymentStatus(status)
	return &payment, nil
}

func (r *postgresPaymentRepository) Update(ctx context.Context, payment *entity.Payment) error {
	query := `
		UPDATE payments
		SET status = $2, updated_at = $3
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, payment.ID, string(payment.Status), payment.UpdatedAt)
	if err != nil {
		return pErrors.E(pErrors.Internal, "failed to update payment", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return pErrors.E(pErrors.NotFound, "payment not found", nil)
	}

	return nil
}
