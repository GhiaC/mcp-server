package transport

import (
	"context"
)

// Transport defines the interface for MCP transport layers
type Transport interface {
	// Initialize connects to the MCP server and initializes the connection
	Initialize(ctx context.Context, config map[string]interface{}) error

	// ListTools returns all available tools from the remote MCP server
	ListTools(ctx context.Context) ([]Tool, error)

	// CallTool executes a tool on the remote MCP server
	CallTool(ctx context.Context, name string, arguments map[string]interface{}) (*ToolResponse, error)

	// Close closes the transport connection
	Close() error
}

// Tool represents a tool definition from an MCP server
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// ToolResponse represents the response from a tool call
type ToolResponse struct {
	Content []ContentItem `json:"content"`
}

// ContentItem represents a content item in the tool response
type ContentItem struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// InitializeResponse represents the initialize response from MCP server
type InitializeResponse struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    map[string]interface{} `json:"capabilities"`
	ServerInfo      ServerInfo             `json:"serverInfo"`
}

// ServerInfo contains server information
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ToolsListResponse represents the tools list response
type ToolsListResponse struct {
	Tools []Tool `json:"tools"`
}
