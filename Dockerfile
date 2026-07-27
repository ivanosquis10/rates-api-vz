# Build Stage
FROM golang:alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache ca-certificates tzdata

# Cache Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build static binary with CGO disabled (pure Go SQLite)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/server ./cmd/api

# Production Stage
FROM alpine:3.21

# Install runtime certificates (required for HTTPS web scraping)
RUN apk add --no-cache ca-certificates tzdata \
    && adduser -D -g '' appuser

WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/server /app/server

# Prepare volume directory for SQLite persistent database
RUN mkdir -p /data && chown -R appuser:appuser /data /app

USER appuser

EXPOSE 8080

ENV PORT=8080 \
    DB_PATH=/data/rates.db

ENTRYPOINT ["/app/server"]
