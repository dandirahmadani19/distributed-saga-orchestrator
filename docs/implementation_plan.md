# Distributed Saga Orchestrator — Implementation Plan

## What Is This Project?

This project is a **Saga Orchestrator** — a system that manages many steps across different services.

**Simple Analogy**: Think of ordering food at a restaurant:

1. You order food (Order Service)
2. You pay for it (Payment Service)
3. The kitchen prepares your food (Inventory Service)

If step 2 fails (payment rejected), we need to cancel step 1 (cancel the order). This is called **compensation** — undoing what we already did.

---

## How Services Talk to Each Other

### Two Types of Communication

| Type     | Used For                        | Format                  |
| -------- | ------------------------------- | ----------------------- |
| **gRPC** | Service ↔ Service (internal)    | Protobuf (binary, fast) |
| **REST** | Client ↔ API Gateway (external) | JSON (human readable)   |

**Simple Analogy**:

- **gRPC** is like speaking a secret language that only your team understands — very fast, but outsiders cannot read it.
- **REST** is like speaking plain English — everyone can understand, but slower.

```
┌─────────────────────────────────────────────────────────┐
│                    External Clients                      │
│              (Mobile Apps, Web Browser)                  │
└─────────────────────────┬───────────────────────────────┘
                          │ REST/HTTP (JSON)
                          ▼
                 ┌─────────────────┐
                 │   API Gateway   │
                 │  (Orchestrator) │
                 └────────┬────────┘
                          │ gRPC (Protobuf)
         ┌────────────────┼────────────────┐
         ▼                ▼                ▼
   ┌──────────┐    ┌──────────┐    ┌──────────┐
   │  Order   │    │ Payment  │    │Inventory │
   │ Service  │    │ Service  │    │ Service  │
   └──────────┘    └──────────┘    └──────────┘
```

---

## Folder Structure

```
distributed-saga-orchestrator/
│
├── .github/workflows/          # CI/CD pipelines
│   ├── ci.yml
│   ├── cd.yml
│   └── sonar.yml
│
├── proto/                      # Protobuf definitions (shared)
│   ├── order/
│   │   └── order.proto
│   ├── payment/
│   │   └── payment.proto
│   ├── inventory/
│   │   └── inventory.proto
│   └── common/
│       └── common.proto
│
├── services/                   # Each service is a Go module
│   │
│   ├── orchestrator/           # Main orchestrator service
│   │   ├── cmd/main.go
│   │   ├── internal/
│   │   │   ├── domain/         # Business logic (no dependencies)
│   │   │   │   ├── entity/
│   │   │   │   ├── repository/
│   │   │   │   └── service/
│   │   │   ├── application/    # Use cases
│   │   │   │   ├── usecase/
│   │   │   │   └── dto/
│   │   │   └── infrastructure/ # External things
│   │   │       ├── persistence/
│   │   │       ├── grpc/       # gRPC handlers
│   │   │       └── rest/       # REST handlers (for external)
│   │   ├── migrations/
│   │   ├── go.mod
│   │   └── Dockerfile
│   │
│   ├── order/
│   │   ├── cmd/main.go
│   │   ├── internal/
│   │   │   ├── domain/
│   │   │   ├── application/
│   │   │   └── infrastructure/
│   │   │       ├── persistence/
│   │   │       └── grpc/
│   │   ├── migrations/
│   │   ├── go.mod
│   │   └── Dockerfile
│   │
│   ├── payment/
│   │   └── ... (same structure)
│   │
│   └── inventory/
│       └── ... (same structure)
│
├── shared/                     # Shared code (used by all services)
│   ├── pkg/
│   │   ├── config/
│   │   ├── database/
│   │   ├── logger/
│   │   ├── retry/
│   │   └── grpcutil/           # gRPC helpers
│   └── go.mod
│
├── deployments/
│   └── docker/
│       └── docker-compose.yml
│
├── scripts/
│   ├── generate-proto.sh       # Generate Go code from .proto
│   └── migrate.sh
│
├── docs/
├── sonar-project.properties
├── .golangci.yml
├── buf.yaml                    # Protobuf linter config
├── Makefile
└── README.md
```

---

## Protobuf Definitions

### What is Protobuf?

Protobuf is a way to define data format. It is like a contract between services.

**Simple Analogy**: Protobuf is like a form template. Everyone agrees what fields are needed before they share information.

### Order Service Proto

```protobuf
// proto/order/order.proto
syntax = "proto3";

package order.v1;

option go_package = "github.com/yourusername/saga/gen/order/v1";

// Request to create an order
message CreateOrderRequest {
    string idempotency_key = 1;  // Unique key to prevent duplicates
    string customer_id = 2;
    repeated OrderItem items = 3;
    double total_amount = 4;
}

// One item in the order
message OrderItem {
    string product_id = 1;
    int32 quantity = 2;
    double price = 3;
}

// Response after creating order
message CreateOrderResponse {
    string order_id = 1;
    string status = 2;
    string created_at = 3;
}

// Request to cancel an order
message CancelOrderRequest {
    string idempotency_key = 1;
    string order_id = 2;
}

// Response after cancelling
message CancelOrderResponse {
    string order_id = 1;
    string status = 2;
}

// The Order Service
service OrderService {
    rpc CreateOrder(CreateOrderRequest) returns (CreateOrderResponse);
    rpc CancelOrder(CancelOrderRequest) returns (CancelOrderResponse);
}
```

### Payment Service Proto

```protobuf
// proto/payment/payment.proto
syntax = "proto3";

package payment.v1;

option go_package = "github.com/yourusername/saga/gen/payment/v1";

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

service PaymentService {
    rpc ProcessPayment(ProcessPaymentRequest) returns (ProcessPaymentResponse);
    rpc RefundPayment(RefundPaymentRequest) returns (RefundPaymentResponse);
}
```

### Inventory Service Proto

```protobuf
// proto/inventory/inventory.proto
syntax = "proto3";

package inventory.v1;

option go_package = "github.com/yourusername/saga/gen/inventory/v1";

message ReserveInventoryRequest {
    string idempotency_key = 1;
    string order_id = 2;
    repeated ReserveItem items = 3;
}

message ReserveItem {
    string product_id = 1;
    int32 quantity = 2;
}

message ReserveInventoryResponse {
    string reservation_id = 1;
    string status = 2;
}

message ReleaseInventoryRequest {
    string idempotency_key = 1;
    string order_id = 2;
}

message ReleaseInventoryResponse {
    string status = 1;
}

service InventoryService {
    rpc ReserveInventory(ReserveInventoryRequest) returns (ReserveInventoryResponse);
    rpc ReleaseInventory(ReleaseInventoryRequest) returns (ReleaseInventoryResponse);
}
```

---

## Clean Architecture Explained

### What is Clean Architecture?

Clean Architecture separates code into layers. Each layer has a specific job.

**Simple Analogy**: Think of a company:

- **Domain Layer** = The CEO (makes business decisions, knows nothing about computers)
- **Application Layer** = The Manager (coordinates work between departments)
- **Infrastructure Layer** = The Workers (do the actual work — talk to database, send emails)

### The Three Layers

```
┌─────────────────────────────────────────────────────────┐
│              INFRASTRUCTURE LAYER                        │
│    (Database, HTTP handlers, gRPC handlers, Clients)    │
│                                                          │
│    This layer talks to the outside world                │
├─────────────────────────────────────────────────────────┤
│              APPLICATION LAYER                           │
│         (Use cases, DTOs, Orchestration)                │
│                                                          │
│    This layer coordinates the work                      │
├─────────────────────────────────────────────────────────┤
│                 DOMAIN LAYER                             │
│    (Entities, Value Objects, Repository Interfaces)    │
│                                                          │
│    This layer contains business rules                   │
│    It has NO external dependencies                      │
└─────────────────────────────────────────────────────────┘

IMPORTANT: Arrows point INWARD only!
Infrastructure → Application → Domain
```

### Example: Order Domain

```go
// internal/domain/entity/order.go
package entity

import "time"

// Order represents an order in our system
// This is a "pure" entity — no database code, no HTTP code
type Order struct {
    ID            string
    CustomerID    string
    Items         []OrderItem
    TotalAmount   float64
    Status        OrderStatus
    CreatedAt     time.Time
}

// OrderItem is one item in an order
type OrderItem struct {
    ProductID string
    Quantity  int
    Price     float64
}

// OrderStatus shows the current state of the order
type OrderStatus string

const (
    OrderStatusCreated   OrderStatus = "CREATED"
    OrderStatusConfirmed OrderStatus = "CONFIRMED"
    OrderStatusCancelled OrderStatus = "CANCELLED"
)

// Cancel changes the order status to cancelled
// This is business logic — it belongs in the domain
func (o *Order) Cancel() error {
    if o.Status != OrderStatusCreated {
        return ErrCannotCancelOrder
    }
    o.Status = OrderStatusCancelled
    return nil
}
```

```go
// internal/domain/repository/order_repository.go
package repository

import "context"

// OrderRepository is an INTERFACE (port)
// It defines WHAT we need, not HOW to do it
// The actual implementation is in the infrastructure layer
type OrderRepository interface {
    Create(ctx context.Context, order *entity.Order) error
    GetByID(ctx context.Context, id string) (*entity.Order, error)
    Update(ctx context.Context, order *entity.Order) error
}
```

---

## SOLID Principles

We will use **SOLID principles** in all our code. SOLID makes code easy to change, test, and understand.

### What is SOLID?

SOLID is 5 rules for writing good code:

| Letter | Name                  | Simple Meaning                             |
| ------ | --------------------- | ------------------------------------------ |
| **S**  | Single Responsibility | One class = one job                        |
| **O**  | Open/Closed           | Add new features without changing old code |
| **L**  | Liskov Substitution   | Child classes can replace parent classes   |
| **I**  | Interface Segregation | Small interfaces are better than big ones  |
| **D**  | Dependency Inversion  | Depend on interfaces, not concrete classes |

### S — Single Responsibility Principle

**Rule**: Each class or function should do only ONE thing.

**Simple Analogy**: In a restaurant, the chef only cooks. The waiter only serves. They don't do each other's jobs.

**Bad Code** ❌:

```go
// This function does too many things!
func CreateOrder(req Request) (*Order, error) {
    // Validate input (job 1)
    if req.CustomerID == "" {
        return nil, errors.New("customer required")
    }

    // Save to database (job 2)
    _, err := db.Exec("INSERT INTO orders...")

    // Send email (job 3)
    sendEmail(req.CustomerID, "Order created!")

    return order, nil
}
```

**Good Code** ✅:

```go
// Each function has ONE job

// Job 1: Validate
func (v *OrderValidator) Validate(req CreateOrderRequest) error {
    if req.CustomerID == "" {
        return ErrCustomerRequired
    }
    return nil
}

// Job 2: Save to database
func (r *OrderRepository) Create(ctx context.Context, order *Order) error {
    return r.db.Exec("INSERT INTO orders...")
}

// Job 3: Send notification
func (n *OrderNotifier) NotifyCreated(order *Order) error {
    return n.emailClient.Send(...)
}

// Use case coordinates all jobs
func (uc *CreateOrderUseCase) Execute(ctx context.Context, req CreateOrderRequest) (*Order, error) {
    if err := uc.validator.Validate(req); err != nil {
        return nil, err
    }

    order := // create order from request

    if err := uc.repo.Create(ctx, order); err != nil {
        return nil, err
    }

    uc.notifier.NotifyCreated(order)
    return order, nil
}
```

### O — Open/Closed Principle

**Rule**: Code should be open for extension, but closed for modification.

**Simple Analogy**: A phone charger. You can plug in different phones (extension), but you don't change the charger itself (closed).

**In Our Project**: We will use interfaces. To add new behavior, we add new implementations — we don't change existing code.

```go
// This interface is CLOSED (we won't change it)
type SagaStep interface {
    Execute(ctx context.Context, payload map[string]any) (map[string]any, error)
    Compensate(ctx context.Context, payload map[string]any) error
}

// We can add NEW steps without changing old code (OPEN for extension)
type CreateOrderStep struct { ... }
type ProcessPaymentStep struct { ... }
type ReserveInventoryStep struct { ... }  // New step - just add it!
```

### L — Liskov Substitution Principle

**Rule**: If class B extends class A, you should be able to use B anywhere you use A.

**Simple Analogy**: If someone says "I need a car", any car will work — Toyota, Honda, Tesla. They all drive the same way.

**In Our Project**: All repository implementations can be swapped.

```go
// Interface
type OrderRepository interface {
    Create(ctx context.Context, order *Order) error
    GetByID(ctx context.Context, id string) (*Order, error)
}

// PostgreSQL implementation
type PostgresOrderRepository struct { ... }

// In-memory implementation (for testing)
type InMemoryOrderRepository struct { ... }

// Both work the same way - you can swap them!
func NewCreateOrderUseCase(repo OrderRepository) *CreateOrderUseCase {
    // Can pass PostgresOrderRepository OR InMemoryOrderRepository
    return &CreateOrderUseCase{repo: repo}
}
```

### I — Interface Segregation Principle

**Rule**: Many small interfaces are better than one big interface.

**Simple Analogy**: A TV remote has buttons for TV only. A universal remote has buttons for TV, DVD, and AC. The TV remote is simpler — you don't need buttons you won't use.

**Bad Code** ❌:

```go
// Too big! Not all implementations need all methods
type OrderRepository interface {
    Create(ctx context.Context, order *Order) error
    GetByID(ctx context.Context, id string) (*Order, error)
    Update(ctx context.Context, order *Order) error
    Delete(ctx context.Context, id string) error
    GetByCustomerID(ctx context.Context, customerID string) ([]*Order, error)
    GetByStatus(ctx context.Context, status string) ([]*Order, error)
    CountByDate(ctx context.Context, date time.Time) (int, error)
}
```

**Good Code** ✅:

```go
// Small, focused interfaces
type OrderCreator interface {
    Create(ctx context.Context, order *Order) error
}

type OrderReader interface {
    GetByID(ctx context.Context, id string) (*Order, error)
}

type OrderUpdater interface {
    Update(ctx context.Context, order *Order) error
}

// Use cases only need what they use
type CreateOrderUseCase struct {
    repo OrderCreator  // Only needs Create!
}

type GetOrderUseCase struct {
    repo OrderReader   // Only needs GetByID!
}
```

### D — Dependency Inversion Principle

**Rule**: High-level code should not depend on low-level code. Both should depend on interfaces.

**Simple Analogy**: A lamp plugs into a wall socket (interface). The lamp doesn't care if electricity comes from solar, wind, or coal. It just needs the socket.

**In Our Project**: Use cases depend on interfaces, not concrete implementations.

```go
// DOMAIN LAYER - defines the interface (the "socket")
type OrderRepository interface {
    Create(ctx context.Context, order *Order) error
}

// APPLICATION LAYER - depends on interface, not implementation
type CreateOrderUseCase struct {
    repo OrderRepository  // Interface! Not PostgresOrderRepository
}

func NewCreateOrderUseCase(repo OrderRepository) *CreateOrderUseCase {
    return &CreateOrderUseCase{repo: repo}
}

// INFRASTRUCTURE LAYER - implements the interface
type PostgresOrderRepository struct {
    db *sql.DB
}

func (r *PostgresOrderRepository) Create(ctx context.Context, order *Order) error {
    // Database code here
}

// MAIN.GO - wires everything together
func main() {
    db := connectToDatabase()
    repo := NewPostgresOrderRepository(db)  // Concrete
    useCase := NewCreateOrderUseCase(repo)  // Passed as interface
}
```

---

## Development Workflow

### How I Will Work

For every change, I will:

1. **Explain FIRST** — Tell you what I will create and why
2. **Show the code** — Display the code I want to add
3. **Wait for your approval** — You say "yes" or ask questions
4. **Make the change** — Only after you approve
5. **Suggest commit** — Give you the commit command

### Example Workflow

```
ME: "I will create the Order entity file. This file contains the
    Order struct with its properties. It follows Single Responsibility
    Principle — it only defines what an Order is.

    Here is the code:
    [code block]

    Should I create this file?"

YOU: "Yes" or "What does this line do?"

ME: [Creates the file]
    "Done! You can commit with: git commit -m 'feat(order): add order domain entity'"
```

---

## Commit Strategy

### Why Small Commits?

Small commits make your GitHub history look professional. Recruiters can see your thought process.

**Simple Analogy**: Writing a book:

- Bad: Write the entire book, then publish (one huge commit)
- Good: Write one chapter at a time (small commits)

### Commit Message Format

```
<type>(<scope>): <short description>

[optional body]
[optional footer]
```

**Types:**
| Type | When to Use |
|------|-------------|
| `feat` | New feature |
| `fix` | Bug fix |
| `refactor` | Code change (no new feature, no bug fix) |
| `docs` | Documentation only |
| `test` | Adding tests |
| `chore` | Build, CI, tooling |

### Example Commit History

```
chore(project): initialize go workspace
chore(shared): add logger package with zerolog
chore(shared): add database connection package
chore(shared): add retry package with backoff

feat(proto): add order service protobuf definition
feat(proto): add payment service protobuf definition
feat(proto): add inventory service protobuf definition
chore(proto): add buf configuration and generate script

feat(order): add order domain entity
feat(order): add order repository interface
feat(order): add create order use case
feat(order): add postgres repository implementation
feat(order): add grpc handler
test(order): add unit tests for order domain
test(order): add integration tests

feat(orchestrator): add saga entity and state machine
feat(orchestrator): add saga repository interface
feat(orchestrator): add distributed lock manager
feat(orchestrator): add saga executor
test(orchestrator): add saga execution tests

chore(ci): add github actions workflow
chore(ci): add sonarqube configuration
docs: add README with architecture diagram
```

---

## Implementation Order

I will build this project step by step. Each step is one or two commits.

### Phase 1: Project Setup (5 commits)

| Step | What We Do              | Commit Message                                  |
| ---- | ----------------------- | ----------------------------------------------- |
| 1.1  | Create folder structure | `chore(project): initialize monorepo structure` |
| 1.2  | Setup Go workspace      | `chore(project): add go.work for workspace`     |
| 1.3  | Add shared logger       | `chore(shared): add logger package`             |
| 1.4  | Add shared database     | `chore(shared): add database package`           |
| 1.5  | Add shared retry        | `chore(shared): add retry package`              |

### Phase 2: Protobuf Setup (4 commits)

| Step | What We Do                | Commit Message                                     |
| ---- | ------------------------- | -------------------------------------------------- |
| 2.1  | Add order.proto           | `feat(proto): add order service definition`        |
| 2.2  | Add payment.proto         | `feat(proto): add payment service definition`      |
| 2.3  | Add inventory.proto       | `feat(proto): add inventory service definition`    |
| 2.4  | Add buf config + generate | `chore(proto): add buf config and generate script` |

### Phase 3: Database Setup (3 commits)

| Step | What We Do           | Commit Message                                    |
| ---- | -------------------- | ------------------------------------------------- |
| 3.1  | Add docker-compose   | `chore(deploy): add docker-compose with postgres` |
| 3.2  | Add migrations       | `feat(db): add database migrations`               |
| 3.3  | Add migration script | `chore(scripts): add migration script`            |

### Phase 4: Order Service (8 commits)

| Step | What We Do                     | Commit Message                           |
| ---- | ------------------------------ | ---------------------------------------- |
| 4.1  | Add order entity               | `feat(order): add order domain entity`   |
| 4.2  | Add order repository interface | `feat(order): add repository interface`  |
| 4.3  | Add create order use case      | `feat(order): add create order use case` |
| 4.4  | Add cancel order use case      | `feat(order): add cancel order use case` |
| 4.5  | Add postgres repository        | `feat(order): add postgres repository`   |
| 4.6  | Add gRPC handler               | `feat(order): add grpc handler`          |
| 4.7  | Add unit tests                 | `test(order): add domain unit tests`     |
| 4.8  | Add main.go                    | `feat(order): add service entrypoint`    |

### Phase 5-6: Payment & Inventory (similar to Phase 4)

### Phase 7: Orchestrator (10 commits)

| Step | What We Do           | Commit Message                                 |
| ---- | -------------------- | ---------------------------------------------- |
| 7.1  | Add saga entity      | `feat(orchestrator): add saga entity`          |
| 7.2  | Add saga step entity | `feat(orchestrator): add saga step entity`     |
| 7.3  | Add saga repository  | `feat(orchestrator): add saga repository`      |
| 7.4  | Add lock manager     | `feat(orchestrator): add distributed lock`     |
| 7.5  | Add saga executor    | `feat(orchestrator): add saga executor`        |
| 7.6  | Add compensator      | `feat(orchestrator): add compensation logic`   |
| 7.7  | Add gRPC clients     | `feat(orchestrator): add grpc clients`         |
| 7.8  | Add REST handlers    | `feat(orchestrator): add rest api handlers`    |
| 7.9  | Add unit tests       | `test(orchestrator): add saga execution tests` |
| 7.10 | Add main.go          | `feat(orchestrator): add service entrypoint`   |

### Phase 8-14: Testing, CI/CD, Docs (see Planning Architecture)

---

## CI/CD Pipeline

### What Happens When You Push Code?

```
Push to GitHub
      │
      ▼
┌─────────────────┐
│  1. Lint Check  │ ─── golangci-lint checks code style
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  2. Unit Tests  │ ─── Run tests for each service
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ 3. Integration  │ ─── Run tests with real database
│    Tests        │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ 4. SonarQube    │ ─── Check code quality
│    Analysis     │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  5. Build       │ ─── Build Docker images
│  Docker Images  │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  6. E2E Tests   │ ─── Test full saga flow
└────────┬────────┘
         │
         ▼
   ✅ All Passed!
```

---

## Verification Checklist

After implementation, we verify:

| Check             | How to Verify                                  |
| ----------------- | ---------------------------------------------- |
| Services start    | `docker compose up` works                      |
| Saga succeeds     | Create order → Payment → Inventory → COMPLETED |
| Saga compensates  | Payment fails → Order cancelled → COMPENSATED  |
| Idempotency works | Same request twice → Same response             |
| Distributed lock  | Kill one orchestrator → Another picks up       |
| gRPC works        | `grpcurl` can call services                    |
| Tests pass        | `make test` shows 80%+ coverage                |
| Lint passes       | `make lint` has no errors                      |
| SonarQube passes  | Quality gate is green                          |
| Observability OK  | Grafana shows metrics, logs, and traces        |

---

## Observability Stack

### What You'll Add

| Tool              | Purpose         | URL                    |
| ----------------- | --------------- | ---------------------- |
| Loki + Fluent Bit | Logs            | http://localhost:3100  |
| Prometheus        | Metrics         | http://localhost:9090  |
| Jaeger            | Traces          | http://localhost:16686 |
| Grafana           | Dashboards      | http://localhost:3000  |
| OpenTelemetry     | Instrumentation | -                      |

**Simple Analogy**: Like a car dashboard showing speed (metrics), fuel (logs), and GPS (traces).

**See Planning Architecture Phase 6 for full implementation details.**

---

## Free Deployment with Render.com

### What is Deployment?

**Simple Analogy**: Deployment is like moving your restaurant from your home kitchen to a real building where customers can visit.

Right now, your project only runs on your computer. Deployment puts it online so recruiters can see it working!

### Why Render.com?

| Feature         | What You Get                                          |
| --------------- | ----------------------------------------------------- |
| **Cost**        | 100% FREE (no credit card needed)                     |
| **Databases**   | 4 free PostgreSQL databases                           |
| **Docker**      | Full support for our Docker setup                     |
| **Auto-Deploy** | Push to GitHub → Auto deploys                         |
| **Sleep Mode**  | Services sleep after 15 minutes (wakes in 30 seconds) |

### Step-by-Step Setup

#### Step 1: Create Render Account

1. Go to [render.com](https://render.com)
2. Click "Get Started"
3. Sign up with your GitHub account (this connects automatically)

#### Step 2: Create render.yaml

Create this file in your project root:

```yaml
# render.yaml
services:
  # Orchestrator - Public REST API
  - type: web
    name: saga-orchestrator
    env: docker
    dockerfilePath: ./services/orchestrator/Dockerfile
    dockerContext: .
    healthCheckPath: /health
    envVars:
      - key: PORT
        value: 8080
      - key: DB_HOST
        fromDatabase:
          name: orchestrator-db
          property: host
      - key: DB_PASSWORD
        fromDatabase:
          name: orchestrator-db
          property: password

  # Order Service - Internal gRPC
  - type: pserv
    name: saga-order
    env: docker
    dockerfilePath: ./services/order/Dockerfile
    dockerContext: .

  # Payment Service
  - type: pserv
    name: saga-payment
    env: docker
    dockerfilePath: ./services/payment/Dockerfile
    dockerContext: .

  # Inventory Service
  - type: pserv
    name: saga-inventory
    env: docker
    dockerfilePath: ./services/inventory/Dockerfile
    dockerContext: .

databases:
  - name: orchestrator-db
    databaseName: saga_orchestrator
  - name: order-db
    databaseName: order_service
  - name: payment-db
    databaseName: payment_service
  - name: inventory-db
    databaseName: inventory_service
```

#### Step 3: Push to GitHub

```bash
git add render.yaml
git commit -m "chore(deploy): add render configuration"
git push origin main
```

#### Step 4: Deploy on Render

1. Go to Render Dashboard
2. Click "New" → "Blueprint"
3. Connect your GitHub repository
4. Select your repository
5. Click "Create New Resources"
6. Wait 5-10 minutes for deployment

#### Step 5: Get Your Live URL

After deployment, you get a URL like:

```
https://saga-orchestrator.onrender.com
```

**Add this to your resume!**

### Auto-Deploy Setup

Every push to `main` branch will automatically deploy.

```yaml
# Already in .github/workflows/cd.yml
name: Deploy

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Trigger Render Deploy
        run: |
          curl -X POST ${{ secrets.RENDER_DEPLOY_HOOK }}
```

### Database Migrations

Migrations run automatically on each deploy. We add this to the Dockerfile:

```dockerfile
# In each service's Dockerfile
COPY migrations ./migrations
COPY scripts/migrate.sh ./migrate.sh

# Run migrations then start service
CMD ["sh", "-c", "./migrate.sh && ./app"]
```

### Testing Your Deployed Service

```bash
# Create a saga
curl -X POST https://saga-orchestrator.onrender.com/api/v1/sagas \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: $(uuidgen)" \
  -d '{
    "saga_type": "order_saga",
    "payload": {
      "customer_id": "customer-123",
      "total_amount": 100.00
    }
  }'

# Check saga status
curl https://saga-orchestrator.onrender.com/api/v1/sagas/{saga_id}
```

### Understanding Free Tier Limits

| Limit                          | What It Means                                  | Impact on You                             |
| ------------------------------ | ---------------------------------------------- | ----------------------------------------- |
| **Services sleep after 15min** | If no one uses it for 15 minutes, it turns off | First request takes 30 seconds to wake up |
| **750 hours/month**            | About 1 month of uptime total                  | Perfect for demos                         |
| **Limited CPU/RAM**            | Not as fast as paid                            | Good enough for portfolio                 |

**Simple Analogy**: Like a store that closes when there are no customers (saves electricity), but opens quickly when someone knocks.

### Monitoring Your Deployment

Render Dashboard shows:

- **Logs** — See what's happening (last 7 days)
- **Metrics** — CPU, memory, request count
- **Deploy History** — All your deployments
- **Health Status** — Is the service running?

### Adding to Your Resume

```markdown
## Distributed Saga Orchestrator

Backend microservices project demonstrating saga pattern for distributed transactions

🔗 **Live Demo**: https://saga-orchestrator.onrender.com  
📖 **API Docs**: https://saga-orchestrator.onrender.com/swagger  
💻 **GitHub**: https://github.com/yourusername/saga-orchestrator

**Tech Stack**: Go, gRPC, PostgreSQL, Docker, Clean Architecture, SOLID Principles
```

### Troubleshooting

| Problem                   | Solution                                      |
| ------------------------- | --------------------------------------------- |
| Service won't start       | Check logs in Render Dashboard                |
| Database connection error | Verify environment variables are set          |
| First request is slow     | Normal! Service was sleeping, wait 30 seconds |
| Service crashed           | Render auto-restarts, check logs for errors   |

---

## Next Steps

After you approve this plan, I will start with **Phase 1: Project Setup**.

I will explain each file I create, what it does, and why we need it. I will use simple language and analogies.

Each change will be a small commit with a clear message.
