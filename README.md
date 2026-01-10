# Zero-Touch Provisioning MCP Server

A Model Context Protocol (MCP) server that provides seamless integration with Ubuntu MAAS (Metal as a Service) for zero-touch provisioning of machines and virtual hosts. This server enables AI-powered automation of infrastructure management through comprehensive tools for machines, VM hosts, networking, and templated deployments.

## 🚀 Features

- **Machine Management**: Commission, deploy, test, and manage physical machines in your MAAS environment
- **VM Host Operations**: List VM hosts, query details, and compose new virtual machines with custom specifications
- **Power Management**: Query and control machine power states
- **Template-Based Deployments**: Create and manage Cloud-Init deployment templates (K3s, K8s, nginx, and custom)
- **Network Infrastructure**: Manage fabrics, VLANs, subnets, and IP address ranges
- **Machine Tagging**: Create, update, and query machine tags for organization and filtering
- **OAuth 1.0 Authentication**: Secure communication with MAAS API using OAuth 1.0 with PLAINTEXT signature
- **Multiple Transport Modes**: Support for stdio, HTTP, and SSE transport protocols
- **Structured Logging**: Comprehensive logging with Zap logger for monitoring and debugging
- **Multi-Platform Support**: Pre-built binaries for Windows, Linux, and macOS (both amd64 and ARM64)
- **Docker Support**: Official Docker images for containerized deployments

## 📋 Prerequisites

- Go 1.23.3 or later
- Ubuntu MAAS instance with API access
- Valid MAAS API credentials

## 🛠️ Installation

1. Clone the repository:
```bash
git clone https://github.com/JarcauCristian/ztp-mcp.git
cd ztp-mcp
```

2. Install dependencies:
```bash
go mod tidy
```

3. Build the project:
```bash
go build -o ztp-mcp cmd/main.go
```

## ⚙️ Configuration

Set the following environment variables:

```bash
# Required: MAAS configuration
export MAAS_BASE_URL="https://your-maas-server.com"
export MAAS_API_KEY="consumer_key:token:secret"

# Optional: MCP server configuration
export MCP_TRANSPORT="stdio"  # Options: stdio, http, sse
export MCP_ADDRESS=":8080"    # Required for http/sse modes
```

### MAAS API Key Format

The `MAAS_API_KEY` must be in the format: `consumer_key:token:secret`

You can obtain your MAAS API key from your MAAS web interface under your user preferences.

## 🚀 Usage

### Stdio Mode (Default)
```bash
./ztp-mcp
```

### HTTP Mode
```bash
export MCP_TRANSPORT=http
export MCP_ADDRESS=":8080"
./ztp-mcp
```

### SSE Mode
```bash
export MCP_TRANSPORT=sse
export MCP_ADDRESS=":8080"
./ztp-mcp
```

## 🔧 Available Tools

### Machine Operations

#### `list_machines`
List all available machines in the MAAS environment, with optional filtering by status.

**Parameters:**
- `status` (optional): Filter machines by status. Options:
  - `new`, `commissioning`, `failed_commissioning`, `ready`
  - `deploying`, `deployed`, `releasing`, `failed_deployment`
  - `allocated`, `retired`, `broken`, `recommissioning`
  - `testing`, `failed_testing`, `rescuing`, `disk_erasing`, `failed_disk_erasing`

**Returns:** JSON array of machine objects (protected machines are automatically filtered out)

#### `list_machine`
Get detailed information about a specific machine by its ID.

**Parameters:**
- `id` (required): The machine system ID (6 alphanumeric characters, e.g., "abc123")

**Returns:** Detailed machine object or empty if protected

#### `commission_machine`
Start the commissioning process on a machine to prepare it for deployment.

**Parameters:**
- `id` (required): The machine system ID

**Returns:** Updated machine object with commissioning status

#### `deploy_machine`
Deploy a machine using a specified Cloud-Init template with custom parameters.

**Parameters:**
- `machineId` (required): The machine system ID
- `templateId` (required): The ID of the deployment template (e.g., "cpu_k3s_deployment", "cpu_k8s_deployment", "nginx_server")
- `templateParameters` (required): JSON object with template-specific parameters. Use `{}` for templates with no parameters

**Returns:** Deployment result with machine configuration

#### `test_machine`
Run testing scripts on a machine to validate hardware and software.

**Parameters:**
- `system_id` (required): The machine system ID
- `enable_ssh` (optional): Enable SSH for testing environment (true/false)
- `parameters` (optional): Custom parameters for test scripts
- `testing_scripts` (optional): Comma-separated list of test script names/tags to run

**Returns:** Testing result and status

### Power Management

#### `power_state`
Query the current power state of a machine.

**Parameters:**
- `id` (required): The machine system ID

**Returns:** Current power state (on/off)

#### `change_power_state`
Change the power state of a machine.

**Parameters:**
- `id` (required): The machine system ID
- `state` (required): Boolean - true to power on, false to power off

**Returns:** Updated power state

### VM Host Operations

#### `list_vm_hosts`
Retrieve all VM hosts available in the MAAS environment.

**Returns:** JSON array of all VM host objects with capabilities and utilization

#### `list_vm_host`
Get detailed information about a specific VM host.

**Parameters:**
- `id` (required): The numeric ID of the VM host

**Returns:** Detailed VM host object

#### `compose_vm_host`
Create a new virtual machine on a specified VM host.

**Parameters:**
- `id` (required): ID of the VM host
- `cores` (required): Number of CPU cores for the VM
- `memory` (required): RAM allocation in MiB
- `storage` (required): Storage allocation in GB
- `hostname` (required): Name for the new VM (alphanumeric, dots, and hyphens allowed)

**Returns:** New VM object with configuration details

### Template Management

#### `retrieve_templates`
List all available deployment templates.

**Parameters:**
- `only_ids` (optional): If true, return only template IDs; if false, return full descriptions

**Returns:** Array of template IDs or template descriptions with metadata

#### `retrieve_template_by_id`
Get detailed information about a specific template.

**Parameters:**
- `id` (required): The template ID

**Returns:** Template description including parameters and requirements

#### `retrieve_template_content`
Get the full content (Cloud-Init configuration) of a template.

**Parameters:**
- `id` (required): The template ID

**Returns:** Template file content (YAML/Cloud-Init format)

#### `create_template`
Create a new deployment template with custom Cloud-Init configuration.

**Parameters:**
- Input schema: `GenericTemplate` object with:
  - `id` (required): Unique template identifier
  - `description` (required): Template description and use case
  - `template_yaml` (required): Cloud-Init YAML template content
  - Additional metadata fields

**Returns:** Confirmation of template creation

#### `delete_template`
Delete an existing deployment template.

**Parameters:**
- `id` (required): The template ID to delete

**Returns:** Confirmation of deletion

### Tag Management

#### `read_tags`
List all available machine tags.

**Returns:** Array of all tag objects with names and descriptions

#### `create_tag`
Create a new machine tag for organizing and filtering machines.

**Parameters:**
- `name` (required): Tag name (unique identifier)
- `comment` (required): Description of the tag's purpose
- `definition` (optional): XPATH query for automatic tag assignment based on hardware details
- `kernel_opts` (optional): Kernel options to apply to machines with this tag

**Returns:** Created tag object

#### `read_tag`
Get detailed information about a specific tag.

**Parameters:**
- `name` (required): The tag name

**Returns:** Tag object with definition and associated machines

#### `update_tag`
Modify an existing tag's properties.

**Parameters:**
- `name` (required): The current tag name
- `new_name` (optional): New tag name
- `comment` (optional): Updated description
- `definition` (optional): Updated XPATH definition

**Returns:** Updated tag object

#### `delete_tag`
Delete a machine tag.

**Parameters:**
- `name` (required): The tag name to delete

**Returns:** Deletion confirmation

#### `list_by_tag`
List all machines/resources of a specific type that have a given tag.

**Parameters:**
- `name` (required): The tag name
- `type` (required): Resource type - `nodes`, `devices`, `machines`, `rack_controllers`, or `region_controllers`

**Returns:** Array of resources with the specified tag

### Subnet Management

#### `list_subnets`
List all subnets in the MAAS environment.

**Returns:** Array of all subnet objects with configuration

#### `create_subnet`
Create a new subnet for network provisioning.

**Parameters:**
- `cidr` (required): Network CIDR notation (e.g., "192.168.1.0/24")
- `name` (optional): Human-readable subnet name
- `description` (optional): Subnet description
- `vlan` (optional): VLAN this subnet belongs to
- `fabric` (optional): Fabric for the subnet
- `vid` (optional): VLAN ID
- `space` (optional): Address space
- `gateway_ip` (optional): Gateway IP address
- `dns_servers` (optional): Comma-separated DNS servers
- `managed` (optional): Whether MAAS manages DHCP/DNS (default: true)

**Returns:** Created subnet object

#### `read_subnet`
Get detailed information about a specific subnet.

**Parameters:**
- `id` (required): The subnet ID

**Returns:** Subnet object with full configuration

#### `update_subnet`
Modify subnet configuration.

**Parameters:**
- `id` (required): The subnet ID
- `cidr`, `name`, `description`, `gateway_ip`, `dns_servers`, etc. (optional): Fields to update
- `managed`, `allow_dns`, `allow_proxy` (optional): Boolean configuration options
- `rdns_mode` (optional): Reverse DNS mode (0=Disabled, 1=Enabled, 2=RFC2317)

**Returns:** Updated subnet object

#### `delete_subnet`
Delete a subnet.

**Parameters:**
- `id` (required): The subnet ID

**Returns:** Deletion confirmation

#### `subnet_ip_addresses`
Get summary of IP addresses assigned to a subnet.

**Parameters:**
- `id` (required): The subnet ID
- `with_username` (optional): Include associated usernames (default: true)
- `with_summary` (optional): Include node/BMC/DNS summaries (default: true)

**Returns:** IP address usage summary

#### `subnet_statistics`
Get detailed statistics about subnet IP usage.

**Parameters:**
- `id` (required): The subnet ID
- `include_ranges` (optional): Include detailed range information (default: false)
- `include_suggestions` (optional): Include suggested gateway and ranges (default: false)

**Returns:** Subnet statistics with usage information

#### `subnet_reserved_ip_ranges`
List IP ranges currently reserved in a subnet.

**Parameters:**
- `id` (required): The subnet ID

**Returns:** Array of reserved IP ranges

#### `subnet_unreserved_ip_ranges`
List IP ranges currently unreserved in a subnet.

**Parameters:**
- `id` (required): The subnet ID

**Returns:** Array of unreserved IP ranges

### Fabric Management

#### `list_fabrics`
List all network fabrics in the MAAS environment.

**Returns:** Array of all fabric objects

#### `create_fabric`
Create a new network fabric.

**Parameters:**
- `name` (optional): Fabric name
- `description` (optional): Fabric description
- `class_type` (optional): Fabric class type

**Returns:** Created fabric object

### VLAN Management

#### `list_vlans`
List all VLANs in a specific fabric.

**Parameters:**
- `fabric_id` (required): The fabric ID to list VLANs from

**Returns:** Array of VLAN objects for the fabric

#### `create_vlan`
Create a new VLAN on a fabric.

**Parameters:**
- `fabric_id` (required): The fabric ID
- `vid` (required): VLAN ID (numeric)
- `name` (optional): VLAN name
- `description` (optional): VLAN description

**Returns:** Created VLAN object

## 📁 Project Structure

```
ztp-mcp/
├── .github/
│   └── workflows/
│       └── go.yaml             # CI/CD pipeline configuration
├── cmd/
│   └── main.go                 # Application entry point
├── internal/
│   └── server/
│       ├── maas_client/
│       │   └── maas-client.go  # MAAS API client with OAuth 1.0 support
│       ├── middleware/
│       │   └── middleware.go   # HTTP middleware (logging, auth)
│       ├── parser/
│       │   └── parse.go        # URI parsing utilities
│       ├── registry/
│       │   └── registry.go     # Registry pattern for tool registration
│       ├── templates/
│       │   ├── cpu_k3s_deployment/
│       │   │   ├── description.json
│       │   │   └── template.yaml
│       │   ├── cpu_k8s_deployment/
│       │   │   ├── description.json
│       │   │   └── template.yaml
│       │   ├── nginx_server/
│       │   │   ├── description.json
│       │   │   └── template.yaml
│       │   ├── template/       # Template scaffolding
│       │   ├── executor.go     # Template execution engine
│       │   ├── template.go     # Single template operations
│       │   └── templates.go    # Template management
│       └── tools/
│           ├── fabrics/        # Fabric management tools
│           ├── node_scripts/   # Node script management tools
│           ├── subnets/        # Subnet management tools
│           ├── tags/           # Machine tag tools
│           ├── vlans/          # VLAN management tools
│           ├── tool.go         # MCP tool interface definition
│           ├── machines.go     # Machine management tools
│           ├── power.go        # Power state management tools
│           ├── templates.go    # Template deployment tools
│           └── vm-hosts.go     # VM host management tools
├── go.mod                      # Go module definition
├── go.sum                      # Go module checksums
├── LICENSE                     # MIT License
├── README.md                   # This file
└── AGENTS.md                   # Development guidelines for AI agents
```

## 🔐 Security

- Uses OAuth 1.0 with PLAINTEXT signature method for MAAS API authentication
- Secure credential management through environment variables
- Request timeouts and proper error handling
- No sensitive data stored in code or logs
- Security scanning with Trivy in CI/CD pipeline
- Vulnerability detection for high and critical severity issues

## 🏗️ Build & Release Process

### Local Build

Build the project for your local system:

```bash
go build -o ztp-mcp ./cmd/main.go
```

For optimized release builds:

```bash
CGO_ENABLED=0 go build -ldflags="-s -w" -o ztp-mcp ./cmd/main.go
```

### Multi-Platform Builds

The CI/CD pipeline automatically builds releases for multiple platforms when you push a version tag (e.g., `v1.0.0`):

- **Windows**: amd64, arm64
- **Linux**: amd64, arm64
- **macOS**: amd64 (Intel), arm64 (Apple Silicon)

Pre-built binaries are available in [GitHub Releases](https://github.com/JarcauCristian/ztp-mcp/releases).

### Download Pre-Built Binaries

```bash
# Download from releases
# Example for Linux x64
wget https://github.com/JarcauCristian/ztp-mcp/releases/download/v1.0.0/ztp-mcp-v1.0.0-linux-amd64.tar.gz
tar xzf ztp-mcp-v1.0.0-linux-amd64.tar.gz
chmod +x ztp-mcp-linux-amd64
./ztp-mcp-linux-amd64
```

## 🐳 Docker Support

### Using Docker

Run the server in a Docker container:

```bash
docker run -e MAAS_BASE_URL="https://your-maas-server.com" \
           -e MAAS_API_KEY="consumer_key:token:secret" \
           -e MCP_TRANSPORT="stdio" \
           ghcr.io/JarcauCristian/ztp-mcp:latest
```

### Using Docker Compose

Create a `docker-compose.yml`:

```yaml
version: '3.8'

services:
  ztp-mcp:
    image: ghcr.io/JarcauCristian/ztp-mcp:latest
    environment:
      - MAAS_BASE_URL=https://your-maas-server.com
      - MAAS_API_KEY=consumer_key:token:secret
      - MCP_TRANSPORT=http
      - MCP_ADDRESS=:8080
    ports:
      - "8080:8080"
    restart: unless-stopped
```

Then start the service:

```bash
docker-compose up -d
```

### Building Docker Image Locally

Build a Docker image from the source:

```bash
docker build -t ztp-mcp:latest .
docker run -e MAAS_BASE_URL="https://your-maas-server.com" \
           -e MAAS_API_KEY="consumer_key:token:secret" \
           ztp-mcp:latest
```

### Container Images

Official container images are published to GitHub Container Registry (GHCR):

```
ghcr.io/JarcauCristian/ztp-mcp:latest          # Latest version
ghcr.io/JarcauCristian/ztp-mcp:v1.0.0          # Specific version
ghcr.io/JarcauCristian/ztp-mcp:1.0             # Latest patch for minor version
ghcr.io/JarcauCristian/ztp-mcp:1                # Latest patch for major version
```

Supports multi-platform images for:
- `linux/amd64`
- `linux/arm64`

## 📝 API Response Format

All tools return responses in the following JSON structure:

```json
{
  "Body": "response_content",
  "StatusCode": 200,
  "Headers": {
    "Content-Type": ["application/json"],
    "..."
  }
}
```

## 🤝 Contributing

Contributions are welcome! Please follow these guidelines:

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Follow the code style guidelines in [AGENTS.md](AGENTS.md)
4. Ensure all tests pass and code is properly formatted
5. Commit your changes (`git commit -m 'Add some amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

### Development Setup

For development guidelines, testing commands, and code style requirements, please refer to [AGENTS.md](AGENTS.md).

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🐛 Troubleshooting

### Common Issues

1. **Authentication Errors**: Verify your `MAAS_API_KEY` format and credentials
2. **Connection Issues**: Check your `MAAS_BASE_URL` and network connectivity
3. **Permission Errors**: Ensure your MAAS user has appropriate permissions for the operations you're trying to perform

### Logging

The server uses structured logging with different log levels. Set the log level using:

```bash
export LOG_LEVEL=debug  # Options: debug, info, warn, error
```

## 📞 Support

For issues and questions:
- Create an issue in this repository
- Check existing issues for similar problems
- Provide logs and configuration details when reporting bugs