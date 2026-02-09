# Multi-stage build for AgentForgeEngine testing
FROM golang:1.24-alpine AS builder

# Install dependencies
RUN apk add --no-cache git bash

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
COPY pkg/ ./pkg/
COPY cmd/ ./cmd/
COPY internal/ ./internal/

# Download dependencies
RUN go mod download

# Copy the entire source code
COPY . .

# Build the main binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o afe ./cmd/agentforgeengine

# Test stage
FROM golang:1.24-alpine AS tester

# Install test dependencies
RUN apk add --no-cache git bash

# Set working directory
WORKDIR /app

# Copy source code
COPY . .

# Download dependencies
RUN go mod download



# Final stage
FROM golang:1.24-alpine AS afe-test

# Install runtime dependencies
RUN apk add --no-cache git bash

# Set working directory
WORKDIR /app

# Copy built binaries from builder stage
COPY --from=builder /app/afe /usr/local/bin/afe
COPY --from=builder /app/pkg/ ./pkg/
COPY --from=builder /app/cmd/ ./cmd/
COPY --from=builder /app/internal/ ./internal/
COPY --from=builder /app/go.mod ./go.mod
COPY --from=builder /app/go.sum ./go.sum

# Copy agents directory
COPY agents/ ./agents/

# Copy scripts
COPY scripts/ ./scripts/

# Copy test framework
COPY --from=builder /app/pkg/testing/ ./pkg/testing/

# Make scripts executable
RUN chmod +x scripts/*.sh

# Create test results directory
RUN mkdir -p /tmp/test-results

# Set environment variables
ENV AFE_LOG_LEVEL=info
ENV CGO_ENABLED=0

# Health check
RUN /usr/local/bin/afe --help || echo "afe binary ready"

# Default command
CMD ["/bin/bash"]
