# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mcp-server ./cmd/mcp-server

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/mcp-server .

# Make binary executable
RUN chmod +x ./mcp-server

# Default port for HTTP mode (when enabled)
# NOTE: Transport defaults to stdio for backwards compatibility
# Set MCP_TRANSPORT=http to enable HTTP mode
ENV MCP_PORT=8080

# Expose the HTTP port (used when MCP_TRANSPORT=http)
EXPOSE 8080

# Set the entrypoint
ENTRYPOINT ["./mcp-server"]
