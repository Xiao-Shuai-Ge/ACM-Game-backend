PROTO_DIR=api/proto
GEN_DIR=api/gen/go

proto:
	protoc -I $(PROTO_DIR) \
	  --go_out $(GEN_DIR) --go_opt paths=source_relative \
	  --go-grpc_out $(GEN_DIR) --go-grpc_opt paths=source_relative \
	  --grpc-gateway_out $(GEN_DIR) --grpc-gateway_opt paths=source_relative \
	  $(PROTO_DIR)/**/**/*.proto

run:
	go run ./cmd/platform

