# Stage 1: Build the Go binary
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy module files first for better caching
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy application source code
COPY backend/ .

# Build statically linked binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /app/bin/recess-server \
    ./cmd/server

# Stage 2: Minimal runtime image
FROM alpine:3.21

# Install runtime dependencies (ca-certificates, wget for healthcheck)
RUN apk add --no-cache ca-certificates tzdata wget

# Create non-root user
RUN addgroup -S recess && adduser -S recess -G recess

WORKDIR /home/recess

# Copy binary from builder
COPY --from=builder /app/bin/recess-server /usr/local/bin/recess-server

# Change ownership
USER recess:recess

# Expose server port
EXPOSE 8080

# Configure healthcheck
HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -q --spider http://127.0.0.1:8080/health || exit 1

# Start server
ENTRYPOINT ["/usr/local/bin/recess-server"]
