# Quick Reference Guide

## Build Commands

### Standard Build
```bash
# Unified image supporting both stdio and HTTP modes
docker build -t incidentio-mcp:latest .
```

### Multi-Architecture Build
```bash
# Single command for multiple architectures
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --tag your-registry/incidentio-mcp:v0.0.1 \
  --push \
  .
```

### Separate Architecture Builds + Manifest
```bash
# Build AMD64
docker buildx build \
  --platform linux/amd64 \
  --tag your-registry/incidentio-mcp:v0.0.1-amd64 \
  --push \
  .

# Build ARM64
docker buildx build \
  --platform linux/arm64 \
  --tag your-registry/incidentio-mcp:v0.0.1-arm64 \
  --push \
  .

# Create multi-arch manifest
docker buildx imagetools create \
  --tag your-registry/incidentio-mcp:v0.0.1 \
  your-registry/incidentio-mcp:v0.0.1-amd64 \
  your-registry/incidentio-mcp:v0.0.1-arm64
```

---

## Run Commands

### Development (Local)
```bash
# Basic HTTP server (default mode)
docker run --rm -p 8080:8080 \
  -e INCIDENT_IO_API_KEY="your-api-key" \
  incidentio-mcp:latest

# stdio mode
docker run --rm \
  -e MCP_TRANSPORT_MODE=stdio \
  -e INCIDENT_IO_API_KEY="your-api-key" \
  incidentio-mcp:latest

# With debug logging
docker run --rm -p 8080:8080 \
  -e INCIDENT_IO_API_KEY="your-api-key" \
  -e MCP_DEBUG=1 \
  -e INCIDENT_IO_DEBUG=1 \
  incidentio-mcp:latest
```

### Production
```bash
# Full production configuration
docker run --rm \
  -p 8080:8080 \
  -u 1001:1001 \
  -e MCP_TRANSPORT_MODE=http \
  -e INCIDENT_IO_API_KEY="$(cat /path/to/secret.txt)" \
  --restart unless-stopped \
  your-registry/incidentio-mcp:v0.0.1

# With resource limits
docker run --rm \
  -p 8080:8080 \
  -u 1001:1001 \
  --cpus="0.5" \
  --memory="256m" \
  -e INCIDENT_IO_API_KEY="$(cat /path/to/secret.txt)" \
  your-registry/incidentio-mcp:v0.0.1
```

### Debug/Interactive
```bash
# Interactive shell
docker run --rm -it \
  -p 8080:8080 \
  -e INCIDENT_IO_API_KEY="your-api-key" \
  --entrypoint /bin/bash \
  incidentio-mcp:latest

# Custom entrypoint with debug
docker run --rm \
  -p 8080:8080 \
  -e MCP_DEBUG=1 \
  -e INCIDENT_IO_DEBUG=1 \
  -e INCIDENT_IO_API_KEY="$(cat /tmp/incident_key.txt)" \
  --entrypoint /bin/bash \
  your-registry/incidentio-mcp:v0.0.1 \
  -c "/app/entrypoint.sh"
```

---

## Environment Variables

| Variable | Values | Default | Description |
|----------|--------|---------|-------------|
| `MCP_TRANSPORT_MODE` | `stdio`, `http`, `sse` | `http` | Transport mode |
| `MCP_HTTP_PORT` | `1-65535` | `8080` | HTTP server port |
| `INCIDENT_IO_API_KEY` | `string` | (required) | API authentication |
| `INCIDENT_IO_BASE_URL` | `url` | `https://api.incident.io/v2` | API base URL |
| `MCP_DEBUG` | `0`, `1` | `0` | MCP protocol debug logs |
| `INCIDENT_IO_DEBUG` | `0`, `1` | `0` | API debug logs |

---

## HTTP Endpoints

### SSE Stream
```bash
# Establish persistent connection
curl -N http://localhost:8080/sse?session=my-session
```

### Send Message
```bash
# POST JSON-RPC message
curl -X POST http://localhost:8080/message?session=my-session \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{...}}'
```

### Health Check
```bash
# Check server health
curl http://localhost:8080/health
```

---

## Testing

### Health Check
```bash
curl http://localhost:8080/health
# Expected: {"status":"healthy","sessions":0}
```

### Initialize Protocol
```bash
# Terminal 1: Start SSE stream
curl -N http://localhost:8080/sse?session=test

# Terminal 2: Send initialize
curl -X POST http://localhost:8080/message?session=test \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "initialize",
    "params": {
      "protocolVersion": "2024-11-05",
      "capabilities": {},
      "clientInfo": {"name": "curl-test", "version": "1.0"}
    }
  }'

# Terminal 2: Send initialized notification
curl -X POST http://localhost:8080/message?session=test \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "notifications/initialized"
  }'

# Terminal 2: Call a tool
curl -X POST http://localhost:8080/message?session=test \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/call",
    "params": {
      "name": "list_incidents",
      "arguments": {}
    }
  }'
```

---

## Troubleshooting

### Port in Use
```bash
# Check what's using the port
lsof -i :8080
netstat -tuln | grep 8080

# Use different port
docker run -p 9090:8080 -e MCP_HTTP_PORT=8080 ...
```

### Permission Denied
```bash
# Run as non-root user
docker run -u 1001:1001 ...

# Or as root (dev only)
docker run -u 0:0 ...
```

### Container Won't Start
```bash
# Check logs
docker logs <container-id>

# Run with debug
docker run -e MCP_DEBUG=1 -e INCIDENT_IO_DEBUG=1 ...

# Interactive shell
docker run -it --entrypoint /bin/bash incidentio-mcp:http-server
```

### API Authentication Errors
```bash
# Verify API key format
echo $INCIDENT_IO_API_KEY | wc -c  # Should be >20 characters

# Test API key directly
curl -H "Authorization: Bearer $INCIDENT_IO_API_KEY" \
  https://api.incident.io/v2/incidents

# Check key is passed to container
docker run --rm \
  -e INCIDENT_IO_API_KEY="your-key" \
  --entrypoint /bin/bash \
  incidentio-mcp:http-server \
  -c 'echo "API Key length: ${#INCIDENT_IO_API_KEY}"'
```

---

## Common Patterns

### Docker Compose
```yaml
version: '3.8'
services:
  mcp-http:
    image: incidentio-mcp:latest
    ports:
      - "8080:8080"
    user: "1001:1001"
    environment:
      - MCP_TRANSPORT_MODE=http
      - INCIDENT_IO_API_KEY=${INCIDENT_IO_API_KEY}
    restart: unless-stopped
```

### Kubernetes
```yaml
apiVersion: v1
kind: Pod
spec:
  securityContext:
    runAsUser: 1001
    runAsGroup: 1001
  containers:
  - name: mcp-http
    image: your-registry/incidentio-mcp:v0.0.1
    ports:
    - containerPort: 8080
    env:
    - name: INCIDENT_IO_API_KEY
      valueFrom:
        secretKeyRef:
          name: incident-io-creds
          key: api-key
```

### Nginx Reverse Proxy
```nginx
upstream mcp_backend {
    server localhost:8080;
}

server {
    listen 80;
    server_name mcp.example.com;

    location / {
        proxy_pass http://mcp_backend;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;

        # SSE-specific settings
        proxy_buffering off;
        proxy_read_timeout 86400;
    }
}
```

---

## Performance Tips

1. **Session Reuse**: Use consistent session IDs to reuse connections
2. **Connection Pooling**: Load balance across multiple instances
3. **Resource Limits**: Set appropriate CPU/memory limits
4. **Health Checks**: Configure liveness/readiness probes
5. **Monitoring**: Track session count and response times

---

## Security Checklist

- [ ] Run as non-root user (`-u 1001:1001`)
- [ ] Use secrets management (not inline keys)
- [ ] Enable TLS/HTTPS in production
- [ ] Restrict network access (firewall rules)
- [ ] Set resource limits
- [ ] Enable audit logging
- [ ] Regular security updates
- [ ] Monitor for anomalous activity

---

## Further Reading

- [HTTP_SERVER.md](./HTTP_SERVER.md) - Comprehensive HTTP server guide
- [CONFIGURATION.md](./CONFIGURATION.md) - Configuration details
- [DEPLOYMENT.md](./DEPLOYMENT.md) - Deployment strategies
- [README.md](../README.md) - Project overview
