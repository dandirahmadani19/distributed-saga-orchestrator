# Distributed Saga Orchestrator — Project Planning

## Project Goal

Build a **production-ready backend portfolio project** that demonstrates:

- Distributed transaction handling
- Failure recovery & compensation
- Idempotency & reliability
- Observability, testing, and profiling
- Clean architecture & clean code practices

This project is designed to showcase **senior-level backend engineering skills**.

---

## Definition of Done (DoD)

The project is considered complete when:

- All services can be started with `docker compose up`
- Saga execution and rollback work correctly
- Real database persistence is used
- Unit tests and integration tests are present
- Profiling and debugging tools are enabled
- Documentation is clear and recruiter-friendly
- Health checks and graceful shutdown are implemented
- Configuration and secrets are properly managed

---

## High-Level System Architecture

### Architecture Style

- Saga Orchestration Pattern
- **Distributed Orchestrator** (multiple instances with database locking)
- Stateless services
- Database per service
- **gRPC for internal communication** (service to service)
- **REST for external API** (client to orchestrator)

### Communication Patterns

| From         | To           | Protocol  | Why                    |
| ------------ | ------------ | --------- | ---------------------- |
| Client       | Orchestrator | REST/HTTP | Easy to test with curl |
| Orchestrator | Services     | gRPC      | Fast, strict contract  |
| Service      | Service      | gRPC      | Fast, strict contract  |

### Components Overview

```
                         ┌──────────────────┐
                         │     Client       │
                         └────────┬─────────┘
                                  │
                                  ▼
                    ┌─────────────────────────────┐
                    │   Load Balancer / Gateway   │
                    └─────────────┬───────────────┘
                                  │
         ┌────────────────────────┼────────────────────────┐
         │                        │                        │
         ▼                        ▼                        ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  Orchestrator   │    │  Orchestrator   │    │  Orchestrator   │
│   Instance 1    │    │   Instance 2    │    │   Instance N    │
└────────┬────────┘    └────────┬────────┘    └────────┬────────┘
         │                      │                      │
         └──────────────────────┼──────────────────────┘
                                │
                    ┌───────────┴───────────┐
                    │   PostgreSQL (Saga)   │
                    │   - sagas table       │
                    │   - saga_steps table  │
                    │   - saga_locks table  │
                    └───────────────────────┘
                                │
         ┌──────────────────────┼──────────────────────┐
         │                      │                      │
         ▼                      ▼                      ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  Order Service  │    │ Payment Service │    │Inventory Service│
│   (PostgreSQL)  │    │   (PostgreSQL)  │    │   (PostgreSQL)  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Distributed Orchestrator Design

To support **multiple orchestrator instances** safely:

1. **Saga Locking** — Use database-level locking (`SELECT FOR UPDATE SKIP LOCKED`) to ensure only one orchestrator processes a saga at a time.

2. **Lease-based Ownership** — Each orchestrator acquires a lease (with TTL) before processing a saga-This prevents stale locks from blocking progress.

3. **Heartbeat Mechanism** — Active orchestrators periodically extend their lease. If heartbeat stops, another instance can take over.

---

## Saga State Machine

### Saga States

```
                    ┌─────────┐
                    │ PENDING │
                    └────┬────┘
                         │ start()
                         ▼
                   ┌───────────┐
              ┌────│ EXECUTING │────┐
              │    └───────────┘    │
              │                     │
        success()              failure()
              │                     │
              ▼                     ▼
       ┌───────────┐         ┌─────────────┐
       │ COMPLETED │         │ COMPENSATING│
       └───────────┘         └──────┬──────┘
                                    │
                          ┌─────────┴─────────┐
                          │                   │
                    compensated()        comp_failed()
                          │                   │
                          ▼                   ▼
                   ┌─────────────┐      ┌────────┐
                   │ COMPENSATED │      │ FAILED │
                   └─────────────┘      └────────┘
```

### State Definitions

| State          | Description                                          |
| -------------- | ---------------------------------------------------- |
| `PENDING`      | Saga created, waiting to start                       |
| `EXECUTING`    | Steps are being executed in order                    |
| `COMPLETED`    | All steps succeeded                                  |
| `COMPENSATING` | A step failed, running compensation in reverse order |
| `COMPENSATED`  | All compensations completed successfully             |
| `FAILED`       | Compensation failed, requires manual intervention    |

### Step States

| State                 | Description                 |
| --------------------- | --------------------------- |
| `PENDING`             | Step not yet executed       |
| `EXECUTING`           | Step currently running      |
| `SUCCEEDED`           | Step completed successfully |
| `FAILED`              | Step execution failed       |
| `COMPENSATING`        | Compensation running        |
| `COMPENSATED`         | Compensation completed      |
| `COMPENSATION_FAILED` | Compensation failed         |

---

## Database Schema

### Orchestrator Database

```sql
-- Saga definitions
CREATE TABLE sagas (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    saga_type       VARCHAR(100) NOT NULL,          -- e.g., 'order_saga'
    status          VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    payload         JSONB NOT NULL,                 -- saga input data
    error_message   TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at    TIMESTAMPTZ,

    CONSTRAINT valid_status CHECK (status IN (
        'PENDING', 'EXECUTING', 'COMPLETED',
        'COMPENSATING', 'COMPENSATED', 'FAILED'
    ))
);

-- Individual saga steps
CREATE TABLE saga_steps (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    saga_id         UUID NOT NULL REFERENCES sagas(id) ON DELETE CASCADE,
    step_name       VARCHAR(100) NOT NULL,          -- e.g., 'create_order'
    step_order      INT NOT NULL,                   -- execution order (1, 2, 3...)
    status          VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    idempotency_key UUID NOT NULL,                  -- for retry safety
    request_payload JSONB,
    response_payload JSONB,
    error_message   TEXT,
    executed_at     TIMESTAMPTZ,
    compensated_at  TIMESTAMPTZ,
    retry_count     INT NOT NULL DEFAULT 0,

    CONSTRAINT valid_step_status CHECK (status IN (
        'PENDING', 'EXECUTING', 'SUCCEEDED', 'FAILED',
        'COMPENSATING', 'COMPENSATED', 'COMPENSATION_FAILED'
    )),
    UNIQUE(saga_id, step_order)
);

-- Distributed lock for saga processing
CREATE TABLE saga_locks (
    saga_id         UUID PRIMARY KEY REFERENCES sagas(id) ON DELETE CASCADE,
    owner_id        VARCHAR(100) NOT NULL,          -- orchestrator instance ID
    acquired_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at      TIMESTAMPTZ NOT NULL,           -- lease expiration

    CONSTRAINT valid_lease CHECK (expires_at > acquired_at)
);

-- Idempotency records (to prevent duplicate saga creation)
CREATE TABLE idempotency_keys (
    key             VARCHAR(255) PRIMARY KEY,
    saga_id         UUID NOT NULL REFERENCES sagas(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at      TIMESTAMPTZ NOT NULL            -- cleanup after TTL
);

-- Indexes for performance
CREATE INDEX idx_sagas_status ON sagas(status) WHERE status IN ('PENDING', 'EXECUTING', 'COMPENSATING');
CREATE INDEX idx_saga_locks_expires ON saga_locks(expires_at);
CREATE INDEX idx_idempotency_expires ON idempotency_keys(expires_at);
```

### Service Databases (Example: Order Service)

```sql
CREATE TABLE orders (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    idempotency_key UUID NOT NULL UNIQUE,
    customer_id     UUID NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'CREATED',
    total_amount    DECIMAL(10, 2) NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT valid_order_status CHECK (status IN (
        'CREATED', 'CONFIRMED', 'CANCELLED'
    ))
);
```

---

## Retry & Timeout Strategy

### Per-Step Timeout

| Service           | Timeout | Rationale                   |
| ----------------- | ------- | --------------------------- |
| Order Service     | 5s      | Simple DB operation         |
| Payment Service   | 10s     | External payment gateway    |
| Inventory Service | 5s      | Simple DB operation         |
| Compensation      | 15s     | Allow more time for cleanup |

### Overall Saga Timeout

- **Default**: 60 seconds
- **Action on timeout**: Mark saga as `FAILED`, trigger alert

### Retry Policy

```go
type RetryConfig struct {
    MaxRetries      int           // Default: 3
    InitialInterval time.Duration // Default: 100ms
    MaxInterval     time.Duration // Default: 10s
    Multiplier      float64       // Default: 2.0 (exponential)
    Jitter          float64       // Default: 0.1 (10% randomness)
}
```

### Retry Decision Matrix

| Error Type         | Retry? | Notes                     |
| ------------------ | ------ | ------------------------- |
| Network timeout    | ✅ Yes | Transient failure         |
| Connection refused | ✅ Yes | Service may be restarting |
| 5xx Server Error   | ✅ Yes | Transient server issue    |
| 4xx Client Error   | ❌ No  | Business logic error      |
| Validation Error   | ❌ No  | Invalid request           |
| Duplicate Key      | ❌ No  | Idempotency working       |

### Circuit Breaker (Optional Enhancement)

```go
type CircuitBreakerConfig struct {
    FailureThreshold   int           // Failures before open (default: 5)
    SuccessThreshold   int           // Successes before close (default: 2)
    Timeout            time.Duration // Time in open state (default: 30s)
}
```

---

## Concurrency Handling

### Saga Locking Strategy

```go
// Acquire lock with SKIP LOCKED to avoid blocking
func (r *SagaRepo) AcquireLock(ctx context.Context, sagaID uuid.UUID, ownerID string, leaseDuration time.Duration) error {
    query := `
        INSERT INTO saga_locks (saga_id, owner_id, expires_at)
        VALUES ($1, $2, NOW() + $3::interval)
        ON CONFLICT (saga_id) DO UPDATE
        SET owner_id = $2, acquired_at = NOW(), expires_at = NOW() + $3::interval
        WHERE saga_locks.expires_at < NOW()  -- Only if lease expired
        RETURNING saga_id
    `
    // Returns error if lock already held by another owner
}

// Extend lease (heartbeat)
func (r *SagaRepo) ExtendLease(ctx context.Context, sagaID uuid.UUID, ownerID string, leaseDuration time.Duration) error {
    query := `
        UPDATE saga_locks
        SET expires_at = NOW() + $3::interval
        WHERE saga_id = $1 AND owner_id = $2
    `
}

// Release lock
func (r *SagaRepo) ReleaseLock(ctx context.Context, sagaID uuid.UUID, ownerID string) error {
    query := `DELETE FROM saga_locks WHERE saga_id = $1 AND owner_id = $2`
}
```

### Picking Up Pending Sagas

Each orchestrator instance periodically polls for work:

```go
func (o *Orchestrator) pollPendingSagas(ctx context.Context) {
    query := `
        SELECT s.id FROM sagas s
        LEFT JOIN saga_locks l ON s.id = l.saga_id
        WHERE s.status IN ('PENDING', 'EXECUTING', 'COMPENSATING')
        AND (l.saga_id IS NULL OR l.expires_at < NOW())
        ORDER BY s.created_at ASC
        LIMIT 10
        FOR UPDATE SKIP LOCKED
    `
    // Process each saga
}
```

---

## Core Concepts Applied

- Saga Pattern (Orchestration)
- Compensation-based rollback
- Idempotent operations
- Eventual consistency
- Failure handling and retries
- Distributed locking
- Lease-based ownership
- Observability (logs, traces, metrics)
- Graceful shutdown
- Health checks

---

## Core Principles & Standards

### SOLID Principles

All code will follow **SOLID principles** for better maintainability and testability:

| Principle                     | What It Means                               | Why It Matters                                     |
| ----------------------------- | ------------------------------------------- | -------------------------------------------------- |
| **S** - Single Responsibility | Each class has one job                      | Easy to understand and change                      |
| **O** - Open/Closed           | Open for extension, closed for modification | Add features without breaking old code             |
| **L** - Liskov Substitution   | Subtypes can replace parent types           | Flexible, swappable implementations                |
| **I** - Interface Segregation | Small, focused interfaces                   | Don't force implementations to have unused methods |
| **D** - Dependency Inversion  | Depend on interfaces, not concrete classes  | Easy to test and swap implementations              |

**Example in our project:**

- Use cases depend on repository **interfaces** (Dependency Inversion)
- Each repository handles only one entity (Single Responsibility)
- We can swap PostgreSQL for in-memory repos without changing use cases (Liskov Substitution)

### Development Workflow

For each code change during implementation:

1. **Explain** what will be created and why
2. **Show** the code to be added
3. **Wait** for approval
4. **Implement** only after approval
5. **Commit** with clear message

This ensures you understand every change and maintain control over the codebase.

---

## Implementation Phases

---

## Phase 1 — Project Skeleton & Standards

### Tasks

- Setup monorepo structure
- Initialize Go modules per service
- Setup shared internal libraries
- Configure linting and formatting

### Tools & Standards

- Go modules
- `golangci-lint`
- `go fmt`
- `go vet`
- Pre-commit hooks

### Clean Architecture Per Service

Each service follows Clean Architecture (Hexagonal/Ports & Adapters):

```
services/{service}/
├── cmd/main.go                    # Entrypoint
├── internal/
│   ├── domain/                    # Core business logic (NO dependencies)
│   │   ├── entity/                # Domain models
│   │   ├── valueobject/           # Value objects
│   │   ├── repository/            # Repository interfaces (ports)
│   │   └── service/               # Domain services
│   │
│   ├── application/               # Use cases
│   │   ├── usecase/               # Application services
│   │   └── dto/                   # Data transfer objects
│   │
│   └── infrastructure/            # External implementations (adapters)
│       ├── persistence/           # Database implementations
│       ├── http/                  # HTTP handlers
│       └── client/                # External service clients
│
├── migrations/
├── config/
├── go.mod
└── Dockerfile
```

**Dependency Rule**: Dependencies point INWARD only (Infrastructure → Application → Domain)

### Output

- Clean repository structure
- Consistent coding standard across services
- Clean Architecture layers in each service

---

## Phase 2 — Database & Persistence Layer

### Tasks

- Setup PostgreSQL via Docker
- Database per service (logical separation)
- Schema migration using `golang-migrate`
- Implement repository pattern

### Tables

**Orchestrator DB:**

- `sagas`
- `saga_steps`
- `saga_locks`
- `idempotency_keys`

**Service DBs:**

- `orders`
- `payments`
- `inventories`

### Concepts

- Durability
- Crash recovery
- Persisted saga state
- Distributed locking

---

## Phase 3 — Saga Orchestrator Core Logic

### Responsibilities

- Create and manage saga lifecycle
- Execute steps sequentially
- Persist step results
- Trigger compensation on failure
- Acquire and manage distributed locks
- Implement heartbeat mechanism

### Concepts

- State machine
- Retry with exponential backoff + jitter
- Reverse-order compensation
- Lease-based ownership
- Lock acquisition with SKIP LOCKED

---

## Phase 4 — Service Implementation (with Idempotency)

### Services

- Order Service
- Payment Service
- Inventory Service

### Rules

- Simple business logic
- Real database interaction
- **Idempotent endpoints** (baked in from the start)
- Explicit compensation endpoints
- Proper error responses

### Endpoints

**Order Service:**

- `POST /orders` — Create order (idempotent)
- `POST /orders/{id}/compensate` — Cancel order
- `GET /health` — Health check
- `GET /ready` — Readiness check

**Payment Service:**

- `POST /payments` — Process payment (idempotent)
- `POST /payments/{id}/refund` — Refund payment
- `GET /health` — Health check
- `GET /ready` — Readiness check

**Inventory Service:**

- `POST /inventory/reserve` — Reserve stock (idempotent)
- `POST /inventory/release` — Release reserved stock
- `GET /health` — Health check
- `GET /ready` — Readiness check

### Idempotency Implementation

```go
// Middleware for idempotent requests
func IdempotencyMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        key := r.Header.Get("Idempotency-Key")
        if key == "" {
            http.Error(w, "Idempotency-Key header required", http.StatusBadRequest)
            return
        }

        // Check if already processed
        if result, exists := cache.Get(key); exists {
            // Return cached response
            w.Write(result)
            return
        }

        // Process and cache
        next.ServeHTTP(w, r)
    })
}
```

---

## Phase 5 — Failure Handling & Resilience

### Tasks

- Implement retry logic with backoff
- Handle timeout scenarios
- Implement circuit breaker (optional)
- Test failure scenarios

### Failure Scenarios to Handle

| Scenario            | Expected Behavior                |
| ------------------- | -------------------------------- |
| Network timeout     | Retry with backoff               |
| Service unavailable | Retry, then fail saga            |
| Payment declined    | No retry, trigger compensation   |
| Duplicate request   | Return cached response           |
| Partial execution   | Resume from last successful step |
| Orchestrator crash  | Another instance picks up saga   |

---

## Phase 6 — Observability Stack

### Tech Stack

| Component           | Technology        | Purpose                        |
| ------------------- | ----------------- | ------------------------------ |
| **Logs**            | Loki + Fluent Bit | Centralized log aggregation    |
| **Metrics**         | Prometheus        | Performance metrics collection |
| **Traces**          | Jaeger            | Distributed tracing            |
| **Visualization**   | Grafana           | Unified dashboard              |
| **Instrumentation** | OpenTelemetry     | Standard instrumentation SDK   |

### 1. Logging: Loki + Fluent Bit

- Structured logging with `zerolog`
- OpenTelemetry trace context in logs
- Fluent Bit ships logs to Loki
- Log levels: DEBUG, INFO, WARN, ERROR

```go
log.Info().
    Str("trace_id", spanCtx.TraceID().String()).
    Str("saga_id", sagaID).
    Str("step", stepName).
    Msg("Executing saga step")
```

### 2. Tracing: Jaeger + OpenTelemetry

- OpenTelemetry SDK for instrumentation
- Jaeger for trace storage and visualization
- Trace propagation via gRPC metadata
- Span per saga step with attributes

### 3. Metrics: Prometheus

- `saga_duration_seconds` (histogram by saga_type, status)
- `saga_step_duration_seconds` (histogram by step_name, status)
- `saga_total` (counter by saga_type, status)
- `saga_retry_total` (counter)
- `grpc_request_duration_seconds` (histogram)
- `db_query_duration_seconds` (histogram)

### 4. Visualization: Grafana

**Dashboards:**

- Saga Overview — Success rate, duration, status breakdown
- Service Health — CPU, memory, request rate, error rate
- Distributed Tracing — Trace timeline, service dependency graph

**Access:** http://localhost:3000 (admin/admin)

### 5. Docker Compose Services

```yaml
loki: # http://localhost:3100
fluent-bit: # Log shipper
prometheus: # http://localhost:9090
jaeger: # http://localhost:16686
grafana: # http://localhost:3000
```

### Health & Readiness Probes

```go
// /health - Liveness probe
func HealthHandler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

// /ready - Readiness probe
func ReadinessHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if err := db.PingContext(r.Context()); err != nil {
            w.WriteHeader(http.StatusServiceUnavailable)
            return
        }
        w.WriteHeader(http.StatusOK)
    }
}
```

---

## Phase 7 — Testing Strategy

### Unit Tests

- Saga state transitions
- Compensation logic
- Idempotency behavior
- Retry logic
- Lock acquisition

### Integration Tests

- Orchestrator ↔ Services
- Real database using testcontainers
- End-to-end saga execution

### Failure Injection Tests

- Simulate payment failure → Validate rollback
- Simulate network timeout → Validate retry
- Simulate orchestrator crash → Validate pickup by another instance

### Test Coverage Target

- Unit tests: 80%+
- Integration tests: Critical paths

---

## Phase 8 — Profiling & Debugging

### Profiling

- Go `pprof` endpoints
- CPU and memory profiling
- Allocation analysis
- Goroutine analysis

```go
import _ "net/http/pprof"

// Enable pprof on debug port
go http.ListenAndServe(":6060", nil)
```

### Debugging

- Context cancellation handling
- Timeout tracing
- Panic recovery with stack traces
- Deadlock detection

---

## Phase 9 — Configuration & Secrets

### Configuration Management

- Environment variables
- Config file (YAML) with `viper`
- Sensible defaults

```yaml
# config.yaml
server:
  port: 8080
  read_timeout: 30s
  write_timeout: 30s

database:
  host: ${DB_HOST:localhost}
  port: ${DB_PORT:5432}
  name: saga_orchestrator
  max_conns: 25
  min_conns: 5

saga:
  default_timeout: 60s
  lock_lease_duration: 30s
  poll_interval: 1s

retry:
  max_attempts: 3
  initial_interval: 100ms
  max_interval: 10s
  multiplier: 2.0
```

### Secrets Management

- Database credentials via environment variables
- Docker secrets in production
- Never commit secrets to repository

---

## Phase 10 — Graceful Shutdown

### Implementation

```go
func (s *Server) GracefulShutdown(ctx context.Context) error {
    // 1. Stop accepting new requests
    s.httpServer.SetKeepAlivesEnabled(false)

    // 2. Wait for in-flight sagas to complete (with timeout)
    s.orchestrator.StopProcessing()
    s.orchestrator.WaitForCompletion(ctx)

    // 3. Release all held locks
    s.orchestrator.ReleaseAllLocks()

    // 4. Close database connections
    s.db.Close()

    // 5. Shutdown HTTP server
    return s.httpServer.Shutdown(ctx)
}
```

---

## Phase 11 — Documentation & GitHub Presentation

### README Sections

1. Project Overview
2. Architecture Diagram
3. Saga Flow Explanation
4. Distributed Orchestrator Design
5. Failure Handling
6. Trade-offs & Design Decisions
7. How to Run
8. Testing & Profiling Guide

### Additional Docs

- `docs/architecture.md`
- `docs/saga-flow.md`
- `docs/failure-scenarios.md`
- `docs/api-reference.md`

### Diagrams to Include

- Architecture overview
- Saga state machine
- Happy path sequence diagram
- Compensation flow sequence diagram
- Distributed lock acquisition flow

---

## Phase 12 — Final Polish

### Checklist

- Clean commit history
- Meaningful commit messages
- Pinned repository on GitHub
- No dead code
- Consistent naming
- All tests passing
- Documentation complete
- Docker Compose working

---

## Phase 13 — CI/CD Pipeline

### GitHub Actions Workflows

| Workflow    | Trigger            | Purpose                    |
| ----------- | ------------------ | -------------------------- |
| `ci.yml`    | Push, PR           | Lint, test, build          |
| `cd.yml`    | Push to main, tags | Build & push Docker images |
| `sonar.yml` | Push, PR           | Code quality analysis      |

### CI Pipeline Stages

1. **Lint & Format** — golangci-lint with comprehensive rules
2. **Unit Tests** — Per-service tests with coverage reports
3. **Integration Tests** — Tests against real PostgreSQL
4. **Build Docker Images** — Multi-stage Docker builds
5. **E2E Tests** — Full saga flow with docker-compose

### Quality Gates

- All lints pass
- Unit test coverage ≥ 80%
- Integration tests pass
- Docker images build successfully
- E2E tests pass

---

## Phase 14 — SonarQube Integration

### Purpose

- Code quality analysis
- Security vulnerability detection
- Technical debt tracking
- Coverage visualization

### Quality Metrics

| Metric                     | Threshold |
| -------------------------- | --------- |
| Code Coverage              | ≥ 80%     |
| Duplicated Lines           | ≤ 3%      |
| Maintainability Rating     | A         |
| Reliability Rating         | A         |
| Security Rating            | A         |
| Security Hotspots Reviewed | 100%      |

### Integration

- Runs on every PR and push to main/develop
- Blocks merge if quality gate fails
- Reports visible in PR comments

---

## Phase 15 — Free Deployment (Render.com)

### Why Render.com?

| Feature        | Render               | Others            |
| -------------- | -------------------- | ----------------- |
| Cost           | 100% Free            | Free tier limited |
| PostgreSQL     | 4 free databases     | Usually 1         |
| Docker Support | ✅ Yes               | Often limited     |
| Auto-Deploy    | ✅ From GitHub       | Varies            |
| Sleep Mode     | After 15min inactive | Varies            |
| Wakeup Time    | ~30 seconds          | Varies            |

### Services to Deploy

1. **Orchestrator Service** — REST API + gRPC
2. **Order Service** — gRPC only
3. **Payment Service** — gRPC only
4. **Inventory Service** — gRPC only
5. **4 PostgreSQL Databases** — One per service

### Deployment Configuration

Each service needs a `render.yaml` file:

```yaml
# render.yaml (in project root)
services:
  # Orchestrator (has public REST API)
  - type: web
    name: saga-orchestrator
    env: docker
    dockerfilePath: ./services/orchestrator/Dockerfile
    dockerContext: .
    envVars:
      - key: DB_HOST
        fromDatabase:
          name: orchestrator-db
          property: host
      - key: DB_PORT
        fromDatabase:
          name: orchestrator-db
          property: port
      - key: DB_USER
        fromDatabase:
          name: orchestrator-db
          property: user
      - key: DB_PASSWORD
        fromDatabase:
          name: orchestrator-db
          property: password
      - key: DB_NAME
        fromDatabase:
          name: orchestrator-db
          property: database
      - key: ORDER_SERVICE_URL
        value: saga-order:50051
      - key: PAYMENT_SERVICE_URL
        value: saga-payment:50051
      - key: INVENTORY_SERVICE_URL
        value: saga-inventory:50051
    healthCheckPath: /health

  # Order Service (internal gRPC)
  - type: pserv
    name: saga-order
    env: docker
    dockerfilePath: ./services/order/Dockerfile
    dockerContext: .
    envVars:
      - key: DB_HOST
        fromDatabase:
          name: order-db
          property: host

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
    user: saga_user

  - name: order-db
    databaseName: order_service
    user: order_user

  - name: payment-db
    databaseName: payment_service
    user: payment_user

  - name: inventory-db
    databaseName: inventory_service
    user: inventory_user
```

### Auto-Deploy with GitHub Actions

```yaml
# .github/workflows/deploy.yml
name: Deploy to Render

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Trigger Render Deploy
        run: |
          curl -X POST ${{ secrets.RENDER_DEPLOY_HOOK }}
```

### Database Migrations on Deploy

Add migration script to each service's Dockerfile:

```dockerfile
# Add to Dockerfile
COPY migrations ./migrations
COPY scripts/migrate.sh ./migrate.sh

# Run migrations on startup
CMD ["sh", "-c", "./migrate.sh && ./app"]
```

### Free Tier Limitations

| Limitation                  | Impact                   | Workaround               |
| --------------------------- | ------------------------ | ------------------------ |
| Services sleep after 15min  | First request takes ~30s | Acceptable for portfolio |
| 750 hours/month per service | Not 24/7                 | Use with demo only       |
| Limited CPU/RAM             | Slower performance       | Fine for demo purposes   |

### Demo URL for Resume

After deployment, you get a URL like:

```
https://saga-orchestrator.onrender.com
```

Add this to your resume:

```
🔗 Live Demo: https://saga-orchestrator.onrender.com
📖 API Docs: https://saga-orchestrator.onrender.com/swagger
```

### Monitoring

Render provides:

- Service logs (last 7 days)
- Metrics (CPU, memory, requests)
- Deploy history
- Auto-restart on crash

---

## Production Concepts Checklist

| Concept                  | Status     |
| ------------------------ | ---------- |
| Clean Architecture       | ✅ Planned |
| Distributed Transactions | ✅ Planned |
| Compensation Logic       | ✅ Planned |
| Idempotency              | ✅ Planned |
| Distributed Locking      | ✅ Planned |
| Retry with Backoff       | ✅ Planned |
| gRPC Communication       | ✅ Planned |
| Protobuf Contracts       | ✅ Planned |
| Observability            | ✅ Planned |
| Testing Pyramid          | ✅ Planned |
| Profiling & Debugging    | ✅ Planned |
| Graceful Shutdown        | ✅ Planned |
| Health Checks            | ✅ Planned |
| Configuration Management | ✅ Planned |
| CI/CD Pipeline           | ✅ Planned |
| SonarQube Integration    | ✅ Planned |
| Documentation            | ✅ Planned |
| Trade-off Awareness      | ✅ Planned |

---

## Trade-offs & Design Decisions

### Why gRPC over REST for Internal Communication?

| Approach | Pros                                               | Cons                        |
| -------- | -------------------------------------------------- | --------------------------- |
| gRPC     | Fast (Protobuf), strict contracts, code generation | Harder to debug             |
| REST     | Easy to test (curl), human readable                | Slower, no strict contracts |

**Decision**: Use gRPC for service-to-service (fast, production-standard). Keep REST for external API (easy to test).

### Why Sync Communication over Message Broker?

### Why Database Locking over Redis?

| Approach   | Pros                    | Cons                    |
| ---------- | ----------------------- | ----------------------- |
| PostgreSQL | No extra infra, ACID    | Slightly higher latency |
| Redis      | Fast, built for locking | Extra dependency        |

**Decision**: Use PostgreSQL to minimize infrastructure complexity.

### Why Orchestration over Choreography?

| Approach      | Pros                                  | Cons                            |
| ------------- | ------------------------------------- | ------------------------------- |
| Orchestration | Centralized control, easier debugging | Single point of coordination    |
| Choreography  | Fully decoupled                       | Hard to trace, complex rollback |

**Decision**: Orchestration for better visibility and simpler compensation logic.

---

## Summary

This project simulates a **real-world distributed system** using Saga Orchestration with **distributed orchestrator support**.

The focus is not business features, but **system reliability, consistency, and failure handling**.

It is designed to be easily explainable during interviews and attractive to recruiters.

Key highlights:

- **Clean Architecture** with proper layering per service
- **gRPC** for internal service communication (production standard)
- **Distributed orchestrator** with database-level locking
- **Comprehensive failure handling** with retry and compensation
- **Production-ready patterns** including idempotency, observability, and graceful shutdown
- **CI/CD Pipeline** with GitHub Actions
- **Code Quality** enforcement with SonarQube and golangci-lint
