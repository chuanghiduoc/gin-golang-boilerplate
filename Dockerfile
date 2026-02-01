# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install build dependencies + UPX for binary compression
RUN apk add --no-cache git ca-certificates tzdata upx

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /app/bin/api ./cmd/api

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /app/bin/migrate ./cmd/migrate

# Compress binaries with UPX (reduces size by ~60-70%)
RUN upx --best --lzma /app/bin/api /app/bin/migrate

# Final stage - use scratch for minimal image
FROM scratch

WORKDIR /app

# Copy timezone data and CA certificates from builder
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy binaries
COPY --from=builder /app/bin/api /app/api
COPY --from=builder /app/bin/migrate /app/migrate
COPY --from=builder /app/db/migrations /app/db/migrations

# Set timezone
ENV TZ=UTC

EXPOSE 8080

ENTRYPOINT ["/app/api"]
