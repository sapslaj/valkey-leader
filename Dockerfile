# Build stage
FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o valkey-leader .

# Final stage
FROM --platform=$BUILDPLATFORM debian:trixie-slim

# Install ca-certificates for TLS connections
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

# Create non-root user
RUN useradd --system --create-home --shell /bin/false valkey-leader

# Copy binary from builder stage
COPY --from=builder /app/valkey-leader /usr/local/bin/valkey-leader

# Switch to non-root user
USER valkey-leader

# Set the entrypoint
ENTRYPOINT ["/usr/local/bin/valkey-leader"]
