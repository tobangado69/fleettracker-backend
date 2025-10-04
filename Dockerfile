# Multi-stage build for FleetTracker Pro Backend
# Indonesian Fleet Management SaaS Application

# Build stage
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with Swagger docs
RUN go mod vendor && CGO_ENABLED=0 GOOS=linux go build -mod=vendor -a -installsuffix cgo -o main cmd/server/main.go

# Verify docs are included
RUN ls -la docs/ || echo "Warning: docs directory not found"

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user for security
RUN adduser -D -s /bin/sh fleettracker

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/main .

# Copy configuration files (if exists)
# COPY --from=builder /app/.env.example .env.example

# Create necessary directories
RUN mkdir -p logs uploads

# Change ownership to non-root user
RUN chown -R fleettracker:fleettracker /app

# Switch to non-root user
USER fleettracker

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./main"]
