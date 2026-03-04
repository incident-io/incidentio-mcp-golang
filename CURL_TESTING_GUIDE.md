# MCP Server Local Testing with cURL

This guide provides comprehensive curl commands to test the MCP server locally using HTTP transport.

## Setup

First, start the server in HTTP mode:

```bash
# Set your API key
export INCIDENT_IO_API_KEY="your-api-key-here"

# Start server in HTTP mode
./bin/mcp-server -transport http -port 8080
```

In another terminal, run these curl commands:

## Basic MCP Protocol Tests

### 1. Health Check
```bash
curl -X GET http://localhost:8080/health
```

Expected: `{"status":"ok"}`

### 2. Initialize Connection
```bash
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "initialize",
    "params": {}
  }'
```

Expected: Server info with protocol version and capabilities

### 3. List All Available Tools
```bash
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/list",
    "params": {}
  }'
```

Expected: List of all 40+ available tools

## Incident Management Tests

### 4. List Incidents
```bash
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 3,
    "method": "tools/call",
    "params": {
      "name": "list_incidents",
      "arguments": {
        "page_size": 5
      }
    }
  }'
```

### 5. Get Specific Incident
```bash
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 4,
    "method": "tools/call",
    "params": {
      "name": "get_incident",
      "arguments": {
        "incident_id": "01HQXXX..."
      }
    }
  }'
```

### 6. List Incident Statuses
```bash
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 5,
    "method": "tools/call",
    "params": {
      "name": "list_incident_statuses",
      "arguments": {}
    }
  }'
```

### 7. List Incident Types
```bash
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 6,
    "method": "tools/call",
    "params": {
      "name": "list_incident_types",
      "arguments": {}
    }
  }'
```

## Severity Tests

### 8. List Severities
```bash
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 7,
    "method": "tools/call",
    "params": {
      "name": "list_severities",
      "arguments": {}
    }
  }'
```

### 9. Get Specific Severity
```bash
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 8,
    "method": "tools/call",
    "params": {
      "name": "get_severity",
      "arguments": {
        "severity_id": "01HQXXX..."
      }
    }
  }'
```

## Custom Fields Tests

### 10. List Custom Fields
```bash
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 9,
    "method": "tools/call",
    "params": {
      "name": "list_custom_fields",
      "arguments": {}
    }
  }'
```

### 11. Search Custom Fields
```bash
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 10,
    "method": "tools/call",
    "params": {
      "name": "search_custom_fields",
      "arguments": {
        "query": "priority"
      }
    }
  }'
```

### 12. Get Custom Field
```bash
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 11,
    "method": "tools/call",
    "params": {
      "name": "get_custom_field",
      "arguments": {
        "custom_field_id": "01HQXXX..."
      }
    }
  }'
```

## User & Role Tests

### 13. List Users
```bash
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 12,
    "method": "tools/call",
    "params": {
      "name": "list_users",
      "arguments": {
        "page_size": 10
      }
    }
  }'
```

### 14. List Available Incident Roles
```bash
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 13,
    "method": "tools/call",
    "params": {
      "name": "list_available_incident_roles",
      "arguments": {}
    }
  }'
```

## Workflow Tests

### 15. List Workflows
```bash
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 14,
    "method": "tools/call",
    "params": {
      "name": "list_workflows",
      "arguments": {}
    }
  }'
```

### 16. Get Workflow
```bash
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 15,
    "method": "tools/call",
    "params": {
      "name": "get_workflow",
      "arguments": {
        "workflow_id": "01HQXXX..."
      }
    }
  }'
```

## Alert Tests

### 17. List Alerts
```bash
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 16,
    "method": "tools/call",
    "params": {
      "name": "list_alerts",
      "arguments": {
        "page_size": 10
      }
    }
  }'
```

### 18. List Alert Sources
```bash
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 17,
    "method": "tools/call",
    "params": {
      "name": "list_alert_sources",
      "arguments": {}
    }
  }'
```

### 19. List Alert Routes
```bash
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 18,
    "method": "tools/call",
    "params": {
      "name": "list_alert_routes",
      "arguments": {}
    }
  }'
```

## Catalog Tests

### 20. List Catalog Types
```bash
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 19,
    "method": "tools/call",
    "params": {
      "name": "list_catalog_types",
      "arguments": {}
    }
  }'
```

### 21. List Catalog Entries
```bash
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 20,
    "method": "tools/call",
    "params": {
      "name": "list_catalog_entries",
      "arguments": {
        "catalog_type_id": "01HQXXX..."
      }
    }
  }'
```

## Follow-up Tests

### 22. List Follow-ups
```bash
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 21,
    "method": "tools/call",
    "params": {
      "name": "list_follow_ups",
      "arguments": {
        "incident_id": "01HQXXX..."
      }
    }
  }'
```

## Action Tests

### 23. List Actions
```bash
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 22,
    "method": "tools/call",
    "params": {
      "name": "list_actions",
      "arguments": {
        "incident_id": "01HQXXX..."
      }
    }
  }'
```

## Error Handling Tests

### 24. Invalid Method
```bash
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 23,
    "method": "invalid_method",
    "params": {}
  }'
```

Expected: Error response with code -32601 (Method not found)

### 25. Invalid Tool Name
```bash
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 24,
    "method": "tools/call",
    "params": {
      "name": "nonexistent_tool",
      "arguments": {}
    }
  }'
```

Expected: Error response indicating tool not found

### 26. Missing Required Parameter
```bash
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 25,
    "method": "tools/call",
    "params": {
      "name": "get_incident",
      "arguments": {}
    }
  }'
```

Expected: Error response about missing incident_id

## Tips for Testing

1. **Pretty Print JSON**: Pipe curl output through `jq` for readable JSON:
   ```bash
   curl ... | jq '.'
   ```

2. **Save Response**: Save responses to files for inspection:
   ```bash
   curl ... > response.json
   ```

3. **Verbose Mode**: Use `-v` flag to see request/response headers:
   ```bash
   curl -v -X POST ...
   ```

4. **Test Script**: Create a bash script to run multiple tests:
   ```bash
   #!/bin/bash
   for i in {1..5}; do
     echo "Test $i"
     curl -X POST http://localhost:8080/mcp \
       -H "Content-Type: application/json" \
       -d "{\"jsonrpc\":\"2.0\",\"id\":$i,\"method\":\"tools/list\",\"params\":{}}"
     echo ""
   done
   ```

5. **Check Server Logs**: Watch the server terminal for any errors or warnings

## Complete Tool List

Run test #3 (List All Available Tools) to see all 40+ available tools including:
- Incident management (create, update, close, list)
- Custom fields (CRUD operations)
- Severities, workflows, alerts
- User and role management
- Catalog management
- Follow-ups and actions
- Alert routing and sources