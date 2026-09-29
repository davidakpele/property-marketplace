BINARY_NAME   := api
BUILD_DIR     := ./bin
CMD_PATH      := ./cmd/api
MIGRATIONS    := ./migrations
MODULE        := github.com/davidakpele/property-marketplace

LDFLAGS := -ldflags="-w -s"

.PHONY: all build run test test-integration lint fmt vet tidy \
        docker-up docker-down docker-build migrate-up migrate-down clean help

all: build

build:
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_PATH)

run: build
	$(BUILD_DIR)/$(BINARY_NAME)

test:
	go test -v -race -count=1 ./internal/... ./pkg/...

test-integration:
	go test -v -race -count=1 -tags=integration ./tests/integration/...

test-all: test test-integration

lint:
	golangci-lint run ./...

fmt:
	gofmt -w -s ./cmd ./internal ./pkg ./tests
	goimports -w -local $(MODULE) $(shell find ./cmd ./internal ./pkg ./tests -name "*.go")

vet:
	go vet ./...

tidy:
	go mod tidy

docker-build:
	docker build -f deployments/Dockerfile -t $(BINARY_NAME):latest .

docker-up:
	docker compose -f deployments/docker-compose.yml up -d --build

docker-down:
	docker compose -f deployments/docker-compose.yml down

docker-logs:
	docker compose -f deployments/docker-compose.yml logs -f api

migrate-up:
	migrate -path $(MIGRATIONS) -database "$$DATABASE_URL" up

migrate-down:
	migrate -path $(MIGRATIONS) -database "$$DATABASE_URL" down 1

clean:
	@rm -rf $(BUILD_DIR)

help:
	@echo "Targets:"
	@echo "  build             - Build the binary"
	@echo "  run               - Build and run"
	@echo "  test              - Run unit tests"
	@echo "  test-integration  - Run integration tests"
	@echo "  test-all          - Run all tests"
	@echo "  lint              - Run golangci-lint"
	@echo "  fmt               - Format source files"
	@echo "  vet               - Run go vet"
	@echo "  tidy              - Tidy go modules"
	@echo "  docker-build      - Build Docker image"
	@echo "  docker-up         - Start all services via docker compose"
	@echo "  docker-down       - Stop all services"
	@echo "  docker-logs       - Tail API logs"
	@echo "  migrate-up        - Apply all pending migrations"
	@echo "  migrate-down      - Roll back one migration"
	@echo "  clean             - Remove build artifacts"
