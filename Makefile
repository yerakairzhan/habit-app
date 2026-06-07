.PHONY: all build proto test lint fmt run docker-up docker-down \
        fe-install fe-dev fe-build fe-type-check

BINARY   := server
BUILD_DIR:= ./bin
CMD      := ./cmd/server

# ─── Default ──────────────────────────────────────────────────────────────────
all: proto build

# ─── Build ────────────────────────────────────────────────────────────────────
build:
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY) $(CMD)
	@echo "✓ Built $(BUILD_DIR)/$(BINARY)"

# ─── Proto codegen (uses buf — no local protoc install needed) ────────────────
proto:
	@which buf >/dev/null 2>&1 || \
		(echo "Install buf: https://buf.build/docs/installation" && exit 1)
	@mkdir -p gen
	buf generate
	@echo "✓ Proto generated in gen/"

# ─── Tests ────────────────────────────────────────────────────────────────────
test:
	go test ./... -v -race -count=1

test-cover:
	go test ./... -coverprofile=coverage.out -covermode=atomic
	go tool cover -html=coverage.out -o coverage.html

# ─── Lint ─────────────────────────────────────────────────────────────────────
lint:
	@which golangci-lint >/dev/null 2>&1 || \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run ./...

fmt:
	gofmt -s -w .

vet:
	go vet ./...

# ─── Run locally ──────────────────────────────────────────────────────────────
run:
	@[ -f .env ] && export $$(grep -v '^#' .env | xargs) || true
	go run $(CMD)

# ─── Docker ───────────────────────────────────────────────────────────────────
docker-up:
	docker compose up -d --build
	@echo "✓  Backend: http://localhost:50051"
	@echo "✓  Frontend: http://localhost:5173"

docker-down:
	docker compose down -v

docker-logs:
	docker compose logs -f server

docker-logs-fe:
	docker compose logs -f frontend

# ─── Migrations ───────────────────────────────────────────────────────────────
migrate-up:
	@which migrate >/dev/null 2>&1 || \
		go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	migrate -path ./migrations -database "$${DATABASE_URL}" up

migrate-down:
	migrate -path ./migrations -database "$${DATABASE_URL}" down 1

migrate-create:
	@[ -n "$(name)" ] || (echo "Usage: make migrate-create name=my_migration" && exit 1)
	migrate create -ext sql -dir ./migrations -seq $(name)

# ─── Frontend ─────────────────────────────────────────────────────────────────
fe-install:
	cd frontend && npm install

fe-dev:
	cd frontend && npm run dev

fe-build:
	cd frontend && npm run build

fe-preview:
	cd frontend && npm run preview

fe-type-check:
	cd frontend && npm run type-check

# ─── Quick test with curl (no grpcurl needed) ─────────────────────────────────
curl-register:
	curl -s -X POST http://localhost:50051/habit.v1.AuthService/Register \
		-H "Content-Type: application/json" \
		-H "Connect-Protocol-Version: 1" \
		-d '{"email":"test@example.com","password":"password123","name":"Test User"}' | jq .

curl-login:
	curl -s -X POST http://localhost:50051/habit.v1.AuthService/Login \
		-H "Content-Type: application/json" \
		-H "Connect-Protocol-Version: 1" \
		-d '{"email":"test@example.com","password":"password123"}' | jq .

curl-health:
	curl -s http://localhost:50051/healthz | jq .

# ─── Cleanup ──────────────────────────────────────────────────────────────────
clean:
	rm -rf $(BUILD_DIR) gen/ coverage.out coverage.html
