# Order Service Verification Guide

This guide explains how to run and verify the Order Service locally using Docker and Make.

## 1. Prerequisites

Ensure you have the following installed:

- **Docker** & **Docker Compose**
- **Make**
- **grpcurl** (optional, for testing gRPC endpoints)
  ```bash
  brew install grpcurl
  ```

## 2. Start the Environment

Run the following command to start the PostgreSQL database and the Order Service (with hot reload):

```bash
make start-dev-order
```

This will run `docker compose up -d` for the order service and its database.

## 3. Run Database Migrations

Apply the database schema to the running PostgreSQL instance:

```bash
make migrate-up-order
```

You should see: `✅ Migrations completed`.

## 4. Verify the Service is Running

Check the logs to ensure the service started correctly:

```bash
make logs-dev-order
```

You should see a log message like: `🚀 Order Service started`.

## 5. Test with gRPCurl

You can test the `CreateOrder` endpoint using `grpcurl`.

### Create an Order

```bash
grpcurl -plaintext -d '{
  "idempotency_key": "unique-key-123",
  "customer_id": "cust-001",
  "total_amount": 100.0,
  "items": [
    {
      "product_id": "prod-abc",
      "quantity": 2,
      "price": 50.0
    }
  ]
}' localhost:50051 order.v1.OrderService/CreateOrder
```

**Expected Response:**

```json
{
  "orderId": "some-uuid",
  "status": "CREATED",
  "createdAt": "2023-..."
}
```

### Cancel an Order (Optional)

Once you have an `orderId` from the previous step, you can try to cancel it:

```bash
grpcurl -plaintext -d '{
  "idempotency_key": "unique-key-cancel-123",
  "order_id": "<YOUR_ORDER_ID>"
}' localhost:50051 order.v1.OrderService/CancelOrder
```

## 6. Stop the Environment

When you are done, stop the containers:

```bash
make stop-dev-order
```
