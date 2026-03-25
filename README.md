# incident.io MCP Server

[![CI](https://github.com/incident-io/incidentio-mcp-golang/actions/workflows/ci.yml/badge.svg)](https://github.com/incident-io/incidentio-mcp-golang/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/incident-io/incidentio-mcp-golang)](https://goreportcard.com/report/github.com/incident-io/incidentio-mcp-golang)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://go.dev/dl/)

A GoLang implementation of an MCP (Model Context Protocol) server for incident.io, providing comprehensive tools to interact with the incident.io API. Built following industry-standard Go project layout patterns.

> ⚠️ **Fair warning!** ⚠️  
> We're experimenting a fair bit in the MCP space, so use at your own risk! 🚀 If you find issues, please raise issues in the repository.

## 🚀 Quick Start

```bash
go run github.com/incident-io/incidentio-mcp-golang/cmd/mcp-server@latest
```

## 📋 Features

- ✅ Complete incident.io V2 API coverage including Custom Fields
- ✅ Workflow automation and management
- ✅ Alert routing and event handling
- ✅ Comprehensive test suite
- ✅ MCP protocol compliant
- ✅ Industry-standard Go project structure
- ✅ Clean, layered architecture (client → handlers → server)

## 🤖 Using with Claude

Add to your Claude Desktop configuration:

**macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`  
**Windows**: `%APPDATA%\Claude\claude_desktop_config.json`

```json
{
  "mcpServers": {
    "incidentio": {
      "command": "go",
      "args":[
        "run",
        "github.com/incident-io/incidentio-mcp-golang/cmd/mcp-server@latest"
      ],
      "env": {
        "INCIDENT_IO_API_KEY": "your-api-key"
      }
    }
  }
}
```
Or, if you'd prefer to run everything in Docker:


```shell
# Clone the repository
git clone https://github.com/incident-io/incidentio-mcp-golang.git
cd incidentio-mcp-golang

# Set up environment
cp .env.example .env
# Edit .env and add your incident.io API key

```

```json
{
    "mcpServers": {
      "incidentio": {
        "command": "docker-compose",
        "args": ["-f", "/path/to/docker-compose.yml", "run", "--rm", "-T", "mcp-server"],
        "env": {
          "INCIDENT_IO_API_KEY": "your-api-key"
        }
      }
    }
}
```


## Available Tools

### Incident Management

- `list_incidents` - List incidents with optional filters
- `get_incident` - Get details of a specific incident
- `create_incident` - Create a new incident
- `create_incident_smart` - Create an incident with smart field resolution
- `update_incident` - Update an existing incident
- `close_incident` - Close an incident with proper workflow
- `list_incident_statuses` - List available incident statuses
- `list_incident_types` - List available incident types
- `list_incident_updates` - List updates for an incident
- `get_incident_update` - Get details of a specific incident update
- `create_incident_update` - Post status updates to incidents
- `delete_incident_update` - Delete an incident update

### Follow-up Management

- `list_follow_ups` - List follow-ups with optional filters
- `get_follow_up` - Get details of a specific follow-up

### Postmortem Management

- `list_postmortems` - List postmortem documents with optional filters
- `get_postmortem` - Get metadata for a specific postmortem document
- `get_postmortem_content` - Get the full markdown content of a postmortem

### Alert Management

- `list_alerts` - List alerts with optional filters
- `get_alert` - Get details of a specific alert
- `list_incident_alerts` - List connections between incidents and alerts
- `list_alert_sources` - List available alert sources
- `create_alert_event` - Create an alert event
- `list_alert_routes` - List alert routes
- `get_alert_route` - Get details of a specific alert route
- `create_alert_route` - Create a new alert route
- `update_alert_route` - Update an existing alert route

### Escalations

- `list_escalation_paths` - List escalation paths in your account
- `get_escalation_path` - Get details of a specific escalation path
- `create_escalation_path` - Create a new escalation path
- `update_escalation_path` - Update an existing escalation path
- `destroy_escalation_path` - Delete an escalation path
- `list_escalations` - List escalations
- `get_escalation` - Get details of a specific escalation
- `create_escalation` - Create a new escalation

### Actions

- `list_actions` - List actions with optional filters
- `get_action` - Get details of a specific action

### Workflow & Automation

- `list_workflows` - List available workflows
- `get_workflow` - Get workflow details
- `update_workflow` - Update workflow configuration

### Severities

- `list_severities` - List available severities
- `get_severity` - Get details of a specific severity

### Team & Roles

- `list_users` - List organization users
- `list_available_incident_roles` - List available incident roles
- `assign_incident_role` - Assign roles to users

### Catalog Management

- `list_catalog_types` - List available catalog types
- `list_catalog_entries` - List catalog entries
- `update_catalog_entry` - Update catalog entries

### Custom Fields

- `list_custom_fields` - List all custom fields
- `get_custom_field` - Get details of a specific custom field
- `search_custom_fields` - Search for custom fields by name or type
- `create_custom_field` - Create a new custom field
- `update_custom_field` - Update custom field configuration
- `delete_custom_field` - Delete a custom field
- `list_custom_field_options` - List all custom field options
- `create_custom_field_option` - Add a new option to a select field

## 📝 Example Usage

```bash
# Through Claude or another MCP client
"Show me all active incidents"
"Create a new incident called 'Database performance degradation' with severity high"
"List alerts for incident INC-123"
"Assign John Doe as incident lead for INC-456"
"Update the Payments service catalog entry with new team information"
"Show me all custom fields configured in incident.io"
"Search for custom fields related to 'priority'"
"Create a new custom field called 'Root Cause' with type single_select"
"List all postmortems for incident INC-123"
"Get the full content of the postmortem for our last major incident"
```

## 📚 Documentation

- **[Development Guide](docs/DEVELOPMENT.md)** - Setup, testing, and contribution guidelines
- **[Configuration Guide](docs/CONFIGURATION.md)** - Environment variables and deployment options
- **[Contributing Guide](docs/CONTRIBUTING.md)** - How to contribute to the project
- **[Testing Guide](docs/TESTING.md)** - Testing documentation and best practices
- **[Deployment Guide](docs/DEPLOYMENT.md)** - Deployment instructions and considerations
- **[Code of Conduct](docs/CODE_OF_CONDUCT.md)** - Community guidelines and standards

## 🏗️ Project Structure

This project follows industry-standard Go project layout:

```
├── cmd/mcp-server/       # Main application entry point
├── internal/
│   ├── client/           # incident.io API client (HTTP layer)
│   ├── handlers/         # MCP protocol handlers (adapter layer)
│   └── server/           # MCP server orchestration
├── pkg/mcp/              # Public MCP types
└── docs/                 # Documentation
```

**Design Philosophy:**
- **`internal/client`**: Pure HTTP API client for incident.io - reusable, testable
- **`internal/handlers`**: MCP protocol adapters - converts MCP requests to API calls
- **`internal/server`**: MCP server that orchestrates handlers and manages the protocol

## 🔧 Troubleshooting

### Common Issues

- **404 errors**: Ensure incident IDs are valid and exist in your instance
- **Authentication errors**: Verify your API key is correct and has proper permissions
- **Parameter errors**: All incident-related tools use `incident_id` as the parameter name

### Debug Mode

Enable debug logging by setting environment variables:

```bash
export MCP_DEBUG=1
export INCIDENT_IO_DEBUG=1
./start-mcp-server.sh
```

## 🤝 Contributing

Contributions are welcome! Please see our [Development Guide](docs/DEVELOPMENT.md) for details on setup, testing, and contribution guidelines.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built with the [Model Context Protocol](https://modelcontextprotocol.io/) specification
- Powered by [incident.io](https://incident.io/) API
- Created with assistance from Claude
