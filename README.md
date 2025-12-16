# MCP Go Server

A minimal, standards-compliant Model Context Protocol (MCP) server implementation in Go. This server provides HTTP-based endpoints for MCP communication and demonstrates a clean, extensible architecture using only the Go standard library.

## Features

- ✅ **MCP Protocol Compliant**: Implements MCP protocol version `2024-11-05`
- ✅ **HTTP Transport**: Uses standard HTTP/JSON for communication
- ✅ **Tool System**: Extensible tool architecture with example Echo tool
- ✅ **Standard Library Only**: No external dependencies required
- ✅ **Comprehensive Tests**: Full test coverage with unit and integration tests
- ✅ **Clean Architecture**: Well-structured, readable, and maintainable code

## Requirements

- Go 1.21 or higher
- No external dependencies (uses only Go standard library)
- (Optional) Google PSE API Key and Search Engine ID for web search functionality

## Installation

1. Clone or download this repository:
```bash
git clone <repository-url>
cd mcp-server
```

2. Initialize dependencies (if needed):
```bash
go mod tidy
```

3. Build the project:
```bash
go build .
```

## Usage

### Starting the Server

Run the server:
```bash
go run .
```

The server will start on port `3333` and log the available endpoints:
```
MCP Server starting on port :3333
Endpoints available:
  GET  /initialize
  GET  /tools/list
  POST /tools/call
```

### API Endpoints

#### 1. Initialize

**Endpoint:** `GET /initialize`

**Description:** Initializes the MCP connection and returns server capabilities.

**Response:**
```json
{
  "protocolVersion": "2024-11-05",
  "capabilities": {
    "tools": true
  },
  "serverInfo": {
    "name": "mcp-go",
    "version": "0.1.0"
  }
}
```

**Example:**
```bash
curl http://localhost:3333/initialize
```

#### 2. List Tools

**Endpoint:** `GET /tools/list`

**Description:** Returns a list of available tools with their schemas.

**Response:**
```json
{
  "tools": [
    {
      "name": "echo",
      "description": "Echo back the provided message",
      "inputSchema": {
        "type": "object",
        "properties": {
          "message": {
            "type": "string",
            "description": "The message to echo back"
          }
        },
        "required": ["message"]
      }
    }
  ]
}
```

**Example:**
```bash
curl http://localhost:3333/tools/list
```

#### 3. Call Tool

**Endpoint:** `POST /tools/call`

**Description:** Executes a tool with the provided arguments.

**Request Body:**
```json
{
  "name": "echo",
  "arguments": {
    "message": "Hello, MCP!"
  }
}
```

**Response:**
```json
{
  "content": [
    {
      "type": "text",
      "text": "Hello, MCP!"
    }
  ]
}
```

**Example:**
```bash
curl -X POST http://localhost:3333/tools/call \
  -H "Content-Type: application/json" \
  -d '{"name":"echo","arguments":{"message":"Hello, MCP!"}}'
```

**Error Responses:**
- `404 Not Found`: Tool name not recognized
- `400 Bad Request`: Invalid JSON or missing required arguments
- `405 Method Not Allowed`: Wrong HTTP method used

#### 4. Google PSE Search Tool

**Tool Name:** `google_pse_search`

**Description:** Search the web using Google Programmable Search Engine

**Request Body:**
```json
{
  "name": "google_pse_search",
  "arguments": {
    "query": "Go programming language",
    "num": 5,
    "start": 1
  }
}
```

**Response:**
```json
{
  "content": [
    {
      "type": "text",
      "text": "Found 1000000 results:\n\n1. Go Programming Language\n   URL: https://go.dev/\n   The Go programming language is an open source project...\n\n..."
    }
  ]
}
```

**Arguments:**
- `query` (required): Search query string
- `num` (optional): Number of results (1-10, default: 10)
- `start` (optional): Start index for pagination (default: 1)

**Example:**
```bash
curl -X POST http://localhost:3333/tools/call \
  -H "Content-Type: application/json" \
  -d '{"name":"google_pse_search","arguments":{"query":"MCP protocol","num":5}}'
```

**Note:** Requires Google PSE API key and Search Engine ID configured in `mcp-config.json` or environment variables.

## Project Structure

```
mcp-server/
├── go.mod                    # Go module definition
├── main.go                   # Application entry point
├── README.md                 # This file
├── mcp-config.json.example  # Example configuration file
├── client/                   # MCP client implementation
│   └── client.go           # Client interface and HTTP client
├── config/                   # Configuration management
│   └── config.go           # Config loading and parsing
├── gateway/                  # Gateway for multiple MCP servers
│   └── gateway.go         # Gateway manager
├── server/                   # HTTP server
│   ├── server.go          # HTTP server and endpoint handlers
│   ├── server_test.go     # Unit tests for endpoints
│   └── integration_test.go # Integration tests
├── tools/                    # Tool implementations
│   ├── echo.go            # Echo tool implementation
│   ├── echo_test.go       # Echo tool tests
│   ├── google_pse.go      # Google PSE search tool
│   ├── google_pse_test.go # Google PSE tests
│   └── proxy/             # Proxy tools for remote MCPs
│       ├── cloudflare.go  # Cloudflare proxy example
│       ├── filesystem.go  # File System MCP proxy
│       └── google_pse.go  # Google PSE proxy wrapper
└── transport/               # Transport layer abstraction
    ├── interface.go       # Transport interface
    └── http.go           # HTTP transport implementation
```

## Testing

The project includes comprehensive test coverage:

### Run All Tests
```bash
go test ./...
```

### Run Tests with Verbose Output
```bash
go test ./... -v
```

### Run Tests with Coverage
```bash
go test ./... -cover
```

### Test Coverage
- **tools**: 100% coverage
- **server**: 59.6% coverage

### Test Files

- `server/server_test.go`: Unit tests for all HTTP endpoints
- `server/integration_test.go`: End-to-end workflow tests
- `tools/echo_test.go`: Tool-specific tests

## Connecting to Other MCP Servers

This server includes a **Gateway** system that allows you to connect to and use tools from other MCP servers (like Cloudflare, GitHub, etc.).

### Architecture

The gateway system consists of:

- **`client/`**: MCP client implementation for connecting to remote servers
- **`gateway/`**: Gateway manager for handling multiple MCP connections
- **`transport/`**: Transport layer abstraction (HTTP, SSE, stdio)
- **`config/`**: Configuration management for MCP servers

### Configuration

#### Option 1: Configuration File

Create a `mcp-config.json` file in the project root:

```json
{
  "port": ":3333",
  "google_pse": {
    "api_key": "YOUR_GOOGLE_PSE_API_KEY",
    "search_engine_id": "YOUR_GOOGLE_PSE_SEARCH_ENGINE_ID"
  },
  "servers": [
    {
      "name": "cloudflare",
      "url": "https://api.cloudflare.com/mcp",
      "transport": "http",
      "enabled": true,
      "prefix": "cloudflare:",
      "auth": {
        "Authorization": "Bearer YOUR_API_TOKEN",
        "X-Auth-Email": "your-email@example.com"
      }
    }
  ]
}
```

**Configuration Options:**
- `port`: Server port (default: `":3333"`). Can be `":3333"` or `"3333"` (will be prefixed with `:` automatically)
- `google_pse`: Google Programmable Search Engine configuration
  - `api_key`: Your Google PSE API key
  - `search_engine_id`: Your Google Custom Search Engine ID (CX)
- `servers`: Array of remote MCP server configurations

#### Option 2: Environment Variables

Set environment variables for configuration:

```bash
# Server port (optional, defaults to :3333)
export MCP_PORT="3333"

# Google PSE configuration (optional)
export GOOGLE_PSE_API_KEY="your-api-key"
export GOOGLE_PSE_SEARCH_ENGINE_ID="your-search-engine-id"

# Remote MCP servers (optional)
export MCP_SERVERS='[{"name":"cloudflare","url":"https://api.cloudflare.com/mcp","transport":"http","enabled":true,"prefix":"cloudflare:"}]'
```

**Note:** Configuration file takes precedence over environment variables. If both are provided, the config file values will be used.

### Using Remote Tools

Once configured, remote tools will be automatically available:

1. **List all tools** (local + remote):
```bash
curl http://localhost:3333/tools/list
```

2. **Call a remote tool** (using prefix):
```bash
curl -X POST http://localhost:3333/tools/call \
  -H "Content-Type: application/json" \
  -d '{"name":"cloudflare:list_zones","arguments":{}}'
```

### Gateway Features

- ✅ **Automatic Tool Discovery**: Remote tools are automatically included in `/tools/list`
- ✅ **Tool Prefixing**: Use prefixes (e.g., `cloudflare:`) to avoid naming conflicts
- ✅ **Multiple Servers**: Connect to multiple MCP servers simultaneously
- ✅ **Error Handling**: Graceful handling of connection failures
- ✅ **Transport Abstraction**: Support for HTTP, SSE, and stdio transports

### Proxy Tools

You can also create proxy wrappers for specific MCP servers. Examples:
- `tools/proxy/cloudflare.go` - Cloudflare MCP proxy
- `tools/proxy/filesystem.go` - File System MCP proxy

#### File System MCP Example

The File System MCP proxy provides convenient methods for file operations:

```go
import "mcp-go/tools/proxy"

// Create proxy
fsProxy := proxy.NewFileSystemProxy(gateway)

// Read a file
response, err := fsProxy.ReadFile(ctx, "/path/to/file.txt")

// Write a file
response, err := fsProxy.WriteFile(ctx, "/path/to/file.txt", "content")

// List directory
response, err := fsProxy.ListDirectory(ctx, "/path/to/dir")

// Create directory
response, err := fsProxy.CreateDirectory(ctx, "/path/to/newdir")

// Delete file
response, err := fsProxy.DeleteFile(ctx, "/path/to/file.txt")
```

**Available File System Tools:**
- `filesystem:read_file` - Read file contents
- `filesystem:write_file` - Write content to file
- `filesystem:list_directory` - List files in directory
- `filesystem:create_directory` - Create a new directory
- `filesystem:delete_file` - Delete a file

**Example API Call:**
```bash
# Read a file
curl -X POST http://localhost:3333/tools/call \
  -H "Content-Type: application/json" \
  -d '{"name":"filesystem:read_file","arguments":{"path":"/path/to/file.txt"}}'

# Write a file
curl -X POST http://localhost:3333/tools/call \
  -H "Content-Type: application/json" \
  -d '{"name":"filesystem:write_file","arguments":{"path":"/path/to/file.txt","content":"Hello World"}}'

# List directory
curl -X POST http://localhost:3333/tools/call \
  -H "Content-Type: application/json" \
  -d '{"name":"filesystem:list_directory","arguments":{"path":"/path/to/dir"}}'
```

## Adding New Tools

To add a new local tool:

1. Create a new file in `tools/` directory (e.g., `tools/my_tool.go`):
```go
package tools

import "fmt"

type MyTool struct {
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    InputSchema map[string]interface{} `json:"inputSchema"`
}

func GetMyTool() MyTool {
    return MyTool{
        Name:        "my_tool",
        Description: "Description of my tool",
        InputSchema: map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "param": map[string]interface{}{
                    "type": "string",
                },
            },
            "required": []string{"param"},
        },
    }
}

func CallMyTool(arguments map[string]interface{}) (string, error) {
    // Implementation here
    return "result", nil
}
```

2. Register the tool in `server/server.go`:
   - Add to `handleToolsList()` to include in tool list
   - Add handler in `handleToolsCall()` to execute the tool

3. Add tests in `tools/my_tool_test.go`

## MCP Protocol Compliance

This implementation follows the MCP specification:

- ✅ Protocol version: `2024-11-05`
- ✅ Tool responses use MCP content format: `{"content": [{"type": "text", "text": "..."}]}`
- ✅ Input schemas follow JSON Schema specification
- ✅ Proper HTTP status codes and error handling

## Development

### Code Style
- Follows Go standard formatting (`go fmt`)
- Uses standard library only
- Clean, readable, and well-documented code

### Building
```bash
# Build binary
go build -o mcp-server .

# Run directly
go run .
```

## License

This project is provided as-is for educational and demonstration purposes.

## Contributing

Contributions are welcome! Please ensure:
- Code follows Go best practices
- All tests pass
- New features include tests
- README is updated if needed

## Google PSE Configuration

To use Google Programmable Search Engine (PSE) for web search:

1. **Get your API credentials:**
   - Create a project in [Google Cloud Console](https://console.cloud.google.com/)
   - Enable the Custom Search API
   - Create an API key
   - Create a Custom Search Engine at [Google Programmable Search Engine](https://programmablesearchengine.google.com/)
   - Get your Search Engine ID (CX)

2. **Add to config file (`mcp-config.json`):**
```json
{
  "google_pse": {
    "api_key": "YOUR_API_KEY",
    "search_engine_id": "YOUR_SEARCH_ENGINE_ID"
  }
}
```

3. **Or use environment variables:**
```bash
export GOOGLE_PSE_API_KEY="your-api-key"
export GOOGLE_PSE_SEARCH_ENGINE_ID="your-search-engine-id"
```

4. **Use the search tool:**
```bash
curl -X POST http://localhost:3333/tools/call \
  -H "Content-Type: application/json" \
  -d '{"name":"google_pse_search","arguments":{"query":"Go programming language","num":5}}'
```

## Example Workflow

1. Create configuration file:
```bash
cp mcp-config.json.example mcp-config.json
# Edit mcp-config.json and add your API keys
```

2. Start the server:
```bash
go run .
```

2. Initialize connection:
```bash
curl http://localhost:3333/initialize
```

3. List available tools:
```bash
curl http://localhost:3333/tools/list
```

4. Call the echo tool:
```bash
curl -X POST http://localhost:3333/tools/call \
  -H "Content-Type: application/json" \
  -d '{"name":"echo","arguments":{"message":"Hello from MCP!"}}'
```

5. Call a File System tool (if configured):
```bash
# Read a file
curl -X POST http://localhost:3333/tools/call \
  -H "Content-Type: application/json" \
  -d '{"name":"filesystem:read_file","arguments":{"path":"/path/to/file.txt"}}'

# List directory
curl -X POST http://localhost:3333/tools/call \
  -H "Content-Type: application/json" \
  -d '{"name":"filesystem:list_directory","arguments":{"path":"/tmp"}}'
```

## Troubleshooting

### Port Already in Use
If port 3333 is already in use, configure a different port:

**Option 1:** Set in `mcp-config.json`:
```json
{
  "port": ":8080"
}
```

**Option 2:** Set environment variable:
```bash
export MCP_SERVER_PORT=":8080"
```

**Option 3:** Modify in `server/server.go` (not recommended):
```go
port := ":8080"  // Change to your desired port
```

### Google PSE Not Working
1. Verify API key and Search Engine ID are set correctly:
   - In `mcp-config.json`: Set `google_pse.enabled: true` and provide credentials
   - Or set `GOOGLE_PSE_API_KEY` and `GOOGLE_PSE_SEARCH_ENGINE_ID` environment variables

2. Check API key permissions and quota limits

3. Verify Search Engine ID (CX) is correct for your Google Custom Search Engine

4. Ensure `google_pse.enabled` is set to `true` in config file

### Connection Refused
Ensure the server is running before making requests:
```bash
go run .
```

### JSON Parsing Errors
Ensure request bodies are valid JSON and Content-Type header is set:
```bash
-H "Content-Type: application/json"
```

## References

- [Model Context Protocol Specification](https://modelcontextprotocol.io/)
- [Go Documentation](https://go.dev/doc/)

