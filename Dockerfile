# Build stage - builds both MCP stdio server and HTTP proxy
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the MCP stdio server binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mcp-server ./cmd/mcp-server

# Build the MCP HTTP proxy binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mcp-http-proxy ./mcp-http-proxy.go

# Final stage - supports both stdio and HTTP modes via entrypoint
FROM alpine:latest

# Install runtime dependencies and create non-root user
RUN apk add --no-cache ca-certificates bash && \
    addgroup -g 1001 incidentio && \
    adduser -D -s /bin/bash -u 1001 -G incidentio incidentio

WORKDIR /app

# Copy both binaries from builder stage
COPY --from=builder /app/mcp-server .
COPY --from=builder /app/mcp-http-proxy .

# Copy entrypoint script
COPY entrypoint.sh .

# Make binaries and script executable
RUN chmod +x ./mcp-server ./mcp-http-proxy ./entrypoint.sh

# Expose HTTP port (only used in HTTP mode)
EXPOSE 8080

# Default to HTTP mode, but can be overridden to stdio
ENV MCP_TRANSPORT_MODE=http
ENV MCP_HTTP_PORT=8080

# Use entrypoint script to support both stdio and HTTP modes
ENTRYPOINT ["./entrypoint.sh"]