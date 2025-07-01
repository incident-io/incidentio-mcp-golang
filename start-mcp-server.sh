#!/bin/bash

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Check if API key is set
if [ -z "$INCIDENT_IO_API_KEY" ]; then
    # Don't output errors to stderr when running under Claude
    # The server will handle missing API key gracefully
    :
fi

exec "$SCRIPT_DIR/bin/mcp-server"