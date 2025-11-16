#!/bin/bash
set -e

# MCP Server Entrypoint
# Supports both stdio and HTTP/SSE transport modes

MODE="${MCP_TRANSPORT_MODE:-http}"
PORT="${MCP_HTTP_PORT:-8080}"

case "$MODE" in
  stdio)
    echo "Starting MCP server in stdio mode..." >&2
    exec /app/mcp-server
    ;;
  http|sse)
    echo "Starting MCP server in HTTP/SSE mode on port $PORT..." >&2
    exec /app/mcp-http-proxy --mcp-server /app/mcp-server --port "$PORT"
    ;;
  *)
    echo "Error: Invalid MCP_TRANSPORT_MODE: $MODE" >&2
    echo "Valid modes: stdio, http, sse" >&2
    exit 1
    ;;
esac
