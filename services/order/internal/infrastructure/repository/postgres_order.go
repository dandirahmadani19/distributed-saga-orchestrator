package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	pErrors "github.com/dandirahmadani19/distributed-saga-orchestrator/platform/errors"
	"github.com/dandirahmadani19/distributed-saga-orchestrator/services/order/internal/domain/entity"
	domainRepo "github.com/dandirahmadani19/distributed-saga-orchestrator/services/order/internal/domain/repository"
	"github.com/google/uuid"
)

type postgresOrderRepository struct {
	db *sql.DB
}

func NewPostgresOrderRepository(db *sql.DB) domainRepo.OrderRepository {
	return &postgresOrderRepository{db: db}
}

func (r *postgresOrderRepository) Create(ctx context.Context, order *entity.Order, idempotencyKey string) (*entity.Order, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, pErrors.E(pErrors.Internal, "failed to begin transaction", err)
	}
	defer tx.Rollback()

	// Generate ID
	order.ID = uuid.New().String()

	// Insert order
	query := `
		INSERT INTO orders (id, customer_id, status, total_amount, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err = tx.ExecContext(ctx, query,
		order.ID, order.CustomerID, order.Status, order.TotalAmount,
		order.CreatedAt, order.UpdatedAt,
	)
	if err != nil {
		return nil, pErrors.E(pErrors.Internal, "failed to insert order", err)
	}

	// Insert order items
	for _, item := range order.Items {
		itemQuery := `
			INSERT INTO order_items (order_id, product_id, quantity, price)
			VALUES ($1, $2, $3, $4)
		`
		_, err = tx.ExecContext(ctx, itemQuery,
			order.ID, item.ProductID, item.Quantity, item.Price,
		)
		if err != nil {
			return nil, pErrors.E(pErrors.Internal, "failed to insert order item", err)
		}
	}

	// Store idempotency key
	response, _ := json.Marshal(order)
	idempQuery := `
		INSERT INTO order_idempotency (key, order_id, operation, response, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err = tx.ExecContext(ctx, idempQuery,
		idempotencyKey, order.ID, "CREATE", response,
		time.Now(), time.Now().Add(24*time.Hour),
	)
	if err != nil {
		return nil, pErrors.E(pErrors.Internal, "failed to insert idempotency key", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, pErrors.E(pErrors.Internal, "failed to commit transaction", err)
	}
	return order, nil
}

func (r *postgresOrderRepository) CheckIdempotency(ctx context.Context, key string) (*entity.Order, error) {
	query := `
		SELECT 
			orders.id,
			orders.customer_id,
			orders.status,
			orders.total_amount,
			orders.created_at,
			orders.updated_at
		FROM order_idempotency
		JOIN orders ON order_idempotency.order_id = orders.id
		WHERE order_idempotency.key = $1 AND order_idempotency.expires_at > NOW()
	`
	var order entity.Order
	var status string
	err := r.db.QueryRowContext(ctx, query, key).Scan(
		&order.ID, &order.CustomerID, &status, &order.TotalAmount,
		&order.CreatedAt, &order.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, pErrors.E(pErrors.Internal, "query order", err)
	}

	order.Status = entity.OrderStatus(status)
	return &order, nil
}

func (r *postgresOrderRepository) FindByID(ctx context.Context, id string) (*entity.Order, error) {
	query := `
		SELECT 
			id,
			customer_id,
			status,
			total_amount,
			created_at,
			updated_at
		FROM orders 
		WHERE id = $1
	`
	var order entity.Order
	var status string
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&order.ID, &order.CustomerID, &status, &order.TotalAmount,
		&order.CreatedAt, &order.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, pErrors.E(pErrors.NotFound, "order not found", err)
		}
		return nil, pErrors.E(pErrors.Internal, "query order", err)
	}

	order.Status = entity.OrderStatus(status)

	// Load items
	itemsQuery := `SELECT product_id, quantity, price FROM order_items WHERE order_id = $1`
	rows, err := r.db.QueryContext(ctx, itemsQuery, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, pErrors.E(pErrors.NotFound, "order item not found", err)
		}
		return nil, pErrors.E(pErrors.Internal, "query order item", err)
	}

	defer rows.Close()
	for rows.Next() {
		var item entity.OrderItem
		if err := rows.Scan(&item.ProductID, &item.Quantity, &item.Price); err != nil {
			return nil, pErrors.E(pErrors.Internal, "scan order item", err)
		}
		order.Items = append(order.Items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, pErrors.E(pErrors.Internal, "query order item", err)
	}

	return &order, nil
}

func (r *postgresOrderRepository) Update(ctx context.Context, order *entity.Order) error {
	query := `
		UPDATE orders
		SET status = $1, updated_at = $2
		WHERE id = $3
	`
	_, err := r.db.ExecContext(ctx, query, order.Status, order.UpdatedAt, order.ID)
	if err != nil {
		return pErrors.E(pErrors.Internal, "failed to update order", err)
	}

	return nil
}
