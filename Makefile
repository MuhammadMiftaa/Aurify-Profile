include .env

$(eval export $(shell sed -ne 's/ *#.*$$//; /./ s/=.*$$// p' .env))

api:
	go run ./cmd/api/main.go

worker:
	go run ./cmd/worker/main.go

go:
	@trap 'kill 0' INT TERM EXIT; \
	go run ./cmd/api/main.go & \
	go run ./cmd/worker/main.go & \
	wait

migrate:
	@if [ -z "$(to)" ]; then \
		goose up; \
	else \
		goose up-to $(to); \
	fi

migration:
	@goose create $(name) sql

rollback:
	@if [ -z "$(to)" ]; then \
		goose down; \
	else \
		goose down-to $(to); \
	fi

migration-status:
	@goose status

seeder:
	@goose -dir ./config/db/seeder create $(name) sql

seed:
	@goose -dir ./config/db/seeder -no-versioning up

seed-reset:
	@goose -dir ./config/db/seeder -no-versioning reset

# Proto and gRPC related commands
install-proto-tools:
	@echo "Installing protoc plugins..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo "Please install protoc manually from: https://grpc.io/docs/protoc-installation/"

proto:
	@echo "Generating gRPC code from proto files..."
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/wallet.proto
	@echo "Proto generation completed!"

clean-proto:
	@echo "Cleaning generated proto files..."
	@rm -f proto/wallet/*.pb.go
	@echo "Clean completed!"