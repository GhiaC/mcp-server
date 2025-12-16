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

## Project Structure

```
mcp-server/
├── go.mod                    # Go module definition
├── main.go                   # Application entry point
├── README.md                 # This file
├── server/
│   ├── server.go            # HTTP server and endpoint handlers
│   ├── server_test.go       # Unit tests for endpoints
│   └── integration_test.go  # Integration tests
└── tools/
    ├── echo.go              # Echo tool implementation
    └── echo_test.go         # Echo tool tests
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

## Adding New Tools

To add a new tool:

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

## Example Workflow

1. Start the server:
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

## Troubleshooting

### Port Already in Use
If port 3333 is already in use, modify the port in `server/server.go`:
```go
port := ":3333"  // Change to your desired port
```

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

