# ─── Stage 1: Proto codegen ──────────────────────────────────────────────────
FROM golang:1.24-bookworm AS proto-builder

# Added libprotobuf-dev so standard Google proto files are placed in /usr/include
RUN apt-get update && apt-get install -y --no-install-recommends \
    protobuf-compiler=3.21.* \
    libprotobuf-dev \
    && rm -rf /var/lib/apt/lists/*

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.34.2 && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.4.0

WORKDIR /app
COPY proto/ proto/
COPY go.mod go.sum ./

RUN mkdir -p gen && \
    protoc \
      --go_out=gen \
      --go_opt=paths=source_relative \
      --go-grpc_out=gen \
      --go-grpc_opt=paths=source_relative \
      -I proto \
      -I /usr/include \
      proto/habit/v1/*.proto

# ─── Stage 2: Go build ───────────────────────────────────────────────────────
FROM golang:1.24-bookworm AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY --from=proto-builder /app/gen ./gen

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w -X main.Version=$(git describe --tags --always 2>/dev/null || echo 'dev')" \
    -o /app/server ./cmd/server

# ─── Stage 3: Final minimal image ────────────────────────────────────────────
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

# Copy binary and migrations
COPY --from=builder /app/server /app/server
COPY --from=builder /app/migrations /app/migrations

EXPOSE 50051

# distroless runs as uid 65532 (nonroot) by default.
USER nonroot:nonroot

ENTRYPOINT ["/app/server"]