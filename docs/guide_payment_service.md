# Payment Service Implementation Guide

This guide will help you implement the **Payment Service**. Follow these steps to build the service layer by layer.

## 1. Proto Definition

Ensure `proto/payment/v1/payment.proto` exists and matches the plan.

```protobuf
syntax = "proto3";

package payment.v1;

option go_package = "github.com/dandirahmadani19/distributed-saga-orchestrator/services/payment/gen/proto/payment/v1";

service PaymentService {
    rpc ProcessPayment(ProcessPaymentRequest) returns (ProcessPaymentResponse);
    rpc RefundPayment(RefundPaymentRequest) returns (RefundPaymentResponse);
}

message ProcessPaymentRequest {
    string idempotency_key = 1;
    string order_id = 2;
    string customer_id = 3;
    double amount = 4;
}

message ProcessPaymentResponse {
    string payment_id = 1;
    string status = 2;
}

message RefundPaymentRequest {
    string idempotency_key = 1;
    string payment_id = 2;
}

message RefundPaymentResponse {
    string payment_id = 1;
    string status = 2;
}
```

Run `make proto-payment` to generate the code.

## 2. Domain Layer

### Entity: `internal/domain/entity/payment.go`

```go
package entity

import "time"

type PaymentStatus string

const (
    PaymentStatusPending   PaymentStatus = "PENDING"
    PaymentStatusProcessed PaymentStatus = "PROCESSED"
    PaymentStatusFailed    PaymentStatus = "FAILED"
    PaymentStatusRefunded  PaymentStatus = "REFUNDED"
)

type Payment struct {
    ID            string
    OrderID       string
    CustomerID    string
    Amount        float64
    Status        PaymentStatus
    CreatedAt     time.Time
    UpdatedAt     time.Time
}
```

### Repository Interface: `internal/domain/repository/payment_repository.go`

```go
package repository

import (
    "context"
    "github.com/dandirahmadani19/distributed-saga-orchestrator/services/payment/internal/domain/entity"
)

type PaymentRepository interface {
    Create(ctx context.Context, payment *entity.Payment, idempotencyKey string) error
    GetByID(ctx context.Context, id string) (*entity.Payment, error)
    Update(ctx context.Context, payment *entity.Payment) error
    CheckIdempotency(ctx context.Context, key string) (*entity.Payment, error)
}
```

## 3. Application Layer

### Use Case: `internal/application/usecase/process_payment.go`

```go
package usecase

import (
    "context"
    "time"

    "github.com/dandirahmadani19/distributed-saga-orchestrator/services/payment/internal/domain/entity"
    "github.com/dandirahmadani19/distributed-saga-orchestrator/services/payment/internal/domain/repository"
    "github.com/dandirahmadani19/distributed-saga-orchestrator/shared/pkg/logger"
    "github.com/google/uuid"
)

type ProcessPaymentUseCase struct {
    repo   repository.PaymentRepository
    logger *logger.Logger
}

func NewProcessPaymentUseCase(repo repository.PaymentRepository, log *logger.Logger) *ProcessPaymentUseCase {
    return &ProcessPaymentUseCase{repo: repo, logger: log}
}

func (uc *ProcessPaymentUseCase) Execute(ctx context.Context, orderID, customerID string, amount float64, idempotencyKey string) (*entity.Payment, error) {
    // 1. Check idempotency
    existing, err := uc.repo.CheckIdempotency(ctx, idempotencyKey)
    if err != nil {
        return nil, err
    }
    if existing != nil {
        uc.logger.Info().Str("key", idempotencyKey).Msg("Returning idempotent response")
        return existing, nil
    }

    // 2. Create new payment
    payment := &entity.Payment{
        ID:         uuid.New().String(),
        OrderID:    orderID,
        CustomerID: customerID,
        Amount:     amount,
        Status:     entity.PaymentStatusProcessed, // Assume success for simplicity
        CreatedAt:  time.Now(),
        UpdatedAt:  time.Now(),
    }

    // 3. Save to DB
    if err := uc.repo.Create(ctx, payment, idempotencyKey); err != nil {
        return nil, err
    }

    return payment, nil
}
```

## 4. Infrastructure Layer

### Postgres Repository: `internal/infrastructure/repository/postgres_payment.go`

(Implement similar to Order Repository, but for `payments` table)

### gRPC Handler: `internal/infrastructure/grpc/payment_handler.go`

```go
package grpc

import (
    "context"

    pb "github.com/dandirahmadani19/distributed-saga-orchestrator/services/payment/gen/proto/payment/v1"
    "github.com/dandirahmadani19/distributed-saga-orchestrator/services/payment/internal/application/usecase"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

type PaymentHandler struct {
    pb.UnimplementedPaymentServiceServer
    processUC *usecase.ProcessPaymentUseCase
}

func NewPaymentHandler(processUC *usecase.ProcessPaymentUseCase) *PaymentHandler {
    return &PaymentHandler{processUC: processUC}
}

func (h *PaymentHandler) ProcessPayment(ctx context.Context, req *pb.ProcessPaymentRequest) (*pb.ProcessPaymentResponse, error) {
    payment, err := h.processUC.Execute(ctx, req.OrderId, req.CustomerId, req.Amount, req.IdempotencyKey)
    if err != nil {
        return nil, status.Error(codes.Internal, err.Error())
    }

    return &pb.ProcessPaymentResponse{
        PaymentId: payment.ID,
        Status:    string(payment.Status),
    }, nil
}
```

## 5. Main & Config

Implement `cmd/main.go` and `internal/config/config.go` exactly like **Order Service**, but change the service name to `payment-service` and port to `50052`.

## 6. Run & Verify

1.  Update `docker-compose.dev.yml` to include `payment-service` and `postgres-payment`.
2.  Run `make start-dev-payment`.
3.  Run migrations (create `services/payment/migrations`).
4.  Verify with `grpcurl` on port `50052`.
