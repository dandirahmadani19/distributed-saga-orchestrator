PROTO_DIR := proto
GEN_DIR := gen

PROTOC := protoc
GO_OUT := --go_out=$(GEN_DIR) --go_opt=paths=source_relative
GRPC_OUT := --go-grpc_out=$(GEN_DIR) --go-grpc_opt=paths=source_relative

SERVICES := order payment inventory
VERSIONS := v1

.PHONY: proto clean

## Generate Go + gRPC code from protobuf
proto:
	@echo "🔨 Generating Go code from Protobuf..."
	@mkdir -p $(GEN_DIR)
	@for service in $(SERVICES); do \
		for version in $(VERSIONS); do \
			if ls $(PROTO_DIR)/$$service/$$version/*.proto >/dev/null 2>&1; then \
				echo "→ Generating $$service/$$version proto"; \
				$(PROTOC) \
					$(GO_OUT) \
					$(GRPC_OUT) \
					$(PROTO_DIR)/$$service/$$version/*.proto; \
			fi; \
		done; \
	done
	@echo "✅ Protobuf generation complete!"

## Remove generated files
clean:
	@echo "🧹 Cleaning generated files..."
	@rm -rf $(GEN_DIR)
