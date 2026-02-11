PROTO_DIR := proto
PROTOC := protoc

# Protoc options
GO_OPT := --go_opt=paths=source_relative
GRPC_OPT := --go-grpc_opt=paths=source_relative

MIGRATE_ORDER_URL := postgresql://saga:saga@postgres-order:5432/order_service?sslmode=disable
MIGRATE_PAYMENT_URL := postgresql://saga:saga@postgres-payment:5432/payment_service?sslmode=disable
NETWORK_NAME := saga-orchestrator-network

.PHONY: proto proto-order proto-inventory proto-payment proto-orchestrator update-grpc clean help

help:
	@echo "📋 Available targets:"
	@echo "  make proto            - Generate all proto files"
	@echo "  make proto-order      - Generate order service protos"
	@echo "  make proto-inventory  - Generate inventory service protos"
	@echo "  make proto-payment    - Generate payment service protos"
	@echo "  make update-grpc      - Upgrade gRPC to latest version"
	@echo "  make clean            - Remove all generated files"
	@echo "  make migrate-up-order - Run order service migrations"
	@echo "  make migrate-down-order - Rollback order service migrations"
	@echo "  make migrate-force-order - Force order service migration version"
	@echo "  make migrate-up-payment - Run payment service migrations"
	@echo "  make migrate-down-payment - Rollback payment service migrations"
	@echo "  make migrate-force-payment - Force payment service migration version"
	@echo "  make build-dev-order - Build and start development mode with Air hot reloading"
	@echo "  make start-dev-order - Start development mode with Air hot reloading"
	@echo "  make stop-dev-order - Stop development mode"
	@echo "  make build-dev-payment - Build and start development mode with Air hot reloading"
	@echo "  make start-dev-payment - Start development mode with Air hot reloading"
	@echo "  make stop-dev-payment - Stop development mode"

# Note: Migrations run inside Docker container to access the Docker network
# Database migration commands for order service
migrate-up-order:
	@echo "🔄 Running order service migrations..."
	@docker run --rm --network $(NETWORK_NAME) \
		-v "$(shell pwd)/services/order/migrations:/migrations" \
		migrate/migrate:latest -path=/migrations \
		-database "$(MIGRATE_ORDER_URL)" up
	@echo "✅ Migrations completed"

migrate-down-order:
	@echo "⬇️  Rolling back order service migrations..."
	@docker run --rm --network $(NETWORK_NAME) \
		-v "$(shell pwd)/services/order/migrations:/migrations" \
		migrate/migrate:latest -path=/migrations \
		-database "$(MIGRATE_ORDER_URL)" down 1
	@echo "✅ Rollback completed"

migrate-force-order:
	@echo "⚠️  Forcing migration version..."
	@docker run --rm --network $(NETWORK_NAME) \
		-v "$(shell pwd)/services/order/migrations:/migrations" \
		migrate/migrate:latest -path=/migrations \
		-database "$(MIGRATE_ORDER_URL)" force $(VERSION)
	@echo "✅ Migration version forced to $(VERSION)"

# Database migration commands for payment service
migrate-up-payment:
	@echo "🔄 Running payment service migrations..."
	@docker run --rm --network $(NETWORK_NAME) \
		-v "$(shell pwd)/services/payment/migrations:/migrations" \
		migrate/migrate:latest -path=/migrations \
		-database "$(MIGRATE_PAYMENT_URL)" up
	@echo "✅ Migrations completed"

migrate-down-payment:
	@echo "⬇️  Rolling back payment service migrations..."
	@docker run --rm --network $(NETWORK_NAME) \
		-v "$(shell pwd)/services/payment/migrations:/migrations" \
		migrate/migrate:latest -path=/migrations \
		-database "$(MIGRATE_PAYMENT_URL)" down 1
	@echo "✅ Rollback completed"

migrate-force-payment:
	@echo "⚠️  Forcing migration version..."
	@docker run --rm --network $(NETWORK_NAME) \
		-v "$(shell pwd)/services/payment/migrations:/migrations" \
		migrate/migrate:latest -path=/migrations \
		-database "$(MIGRATE_PAYMENT_URL)" force $(VERSION)
	@echo "✅ Migration version forced to $(VERSION)"

## Generate all service protos
proto: proto-order proto-inventory proto-payment
	@echo "✅ All proto generation complete!"

## Generate order service protos
proto-order:
	@echo "🔨 Generating order service protos..."
	@mkdir -p services/order/gen
	@$(PROTOC) \
		--go_out=services/order/gen $(GO_OPT) \
		--go-grpc_out=services/order/gen $(GRPC_OPT) \
		$(PROTO_DIR)/order/v1/*.proto
	@echo "✅ Order protos generated"

## Generate inventory service protos
proto-inventory:
	@echo "🔨 Generating inventory service protos..."
	@mkdir -p services/inventory/gen
	@$(PROTOC) \
		--go_out=services/inventory/gen $(GO_OPT) \
		--go-grpc_out=services/inventory/gen $(GRPC_OPT) \
		$(PROTO_DIR)/inventory/v1/*.proto
	@echo "✅ Inventory protos generated"

## Generate payment service protos
proto-payment:
	@echo "🔨 Generating payment service protos..."
	@mkdir -p services/payment/gen
	@$(PROTOC) \
		--go_out=services/payment/gen $(GO_OPT) \
		--go-grpc_out=services/payment/gen $(GRPC_OPT) \
		$(PROTO_DIR)/payment/v1/*.proto
	@echo "✅ Payment protos generated"

## Upgrade gRPC to latest version
update-grpc:
	@echo "⬆️  Upgrading gRPC dependencies..."
	@cd services/order && go get -u google.golang.org/grpc@latest google.golang.org/protobuf@latest
# 	@cd services/inventory && go get -u google.golang.org/grpc@latest google.golang.org/protobuf@latest
	@cd services/payment && go get -u google.golang.org/grpc@latest google.golang.org/protobuf@latest
	@echo "✅ gRPC upgraded in all services"

## Clean generated files
clean:
	@echo "🧹 Cleaning generated files..."
	@rm -rf services/order/gen
	@rm -rf services/inventory/gen
	@rm -rf services/payment/gen
	@rm -rf gen
	@echo "✅ Cleanup complete"

# Development Order Service commands with Air hot reloading
build-dev-order:
	@echo "🔥 Building and starting development mode with Air hot reloading..."
	@docker compose -f deployments/docker/docker-compose.dev.order.yml up -d postgres-order
	@docker compose -f deployments/docker/docker-compose.dev.order.yml up --build -d order-service
	@echo "✅ Development service is running with hot reload"

start-dev-order:
	@echo "🔥 Starting development mode with Air hot reloading..."
	@docker compose -f deployments/docker/docker-compose.dev.order.yml up -d postgres-order
	@docker compose -f deployments/docker/docker-compose.dev.order.yml up -d order-service
	@echo "✅ Development service is running with hot reload"

stop-dev-order:
	@echo "Stopping development service..."
	@docker compose -f deployments/docker/docker-compose.dev.order.yml down postgres-order
	@docker compose -f deployments/docker/docker-compose.dev.order.yml down order-service
	@echo "✅ Development service is stopped"

logs-dev-order:
	@docker compose -f deployments/docker/docker-compose.dev.order.yml logs -f order-service


# Development Payment Service commands with Air hot reloading
build-dev-payment:
	@echo "🔥 Building and starting development mode with Air hot reloading..."
	@docker compose -f deployments/docker/docker-compose.dev.payment.yml up -d postgres-payment
	@docker compose -f deployments/docker/docker-compose.dev.payment.yml up --build -d payment-service
	@echo "✅ Development service is running with hot reload"

start-dev-payment:
	@echo "🔥 Starting development mode with Air hot reloading..."
	@docker compose -f deployments/docker/docker-compose.dev.payment.yml up -d postgres-payment
	@docker compose -f deployments/docker/docker-compose.dev.payment.yml up -d payment-service
	@echo "✅ Development service is running with hot reload"

stop-dev-payment:
	@echo "Stopping development service..."
	@docker compose -f deployments/docker/docker-compose.dev.payment.yml down postgres-payment
	@docker compose -f deployments/docker/docker-compose.dev.payment.yml down payment-service
	@echo "✅ Development service is stopped"

logs-dev-payment:
	@docker compose -f deployments/docker/docker-compose.dev.payment.yml logs -f payment-service

