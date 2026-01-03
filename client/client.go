package client

import (
	"context"
	"fmt"
	"mcp-go/config"
	"mcp-go/transport"
	"sync"
)

// Client represents an MCP client that can connect to remote MCP servers
type Client interface {
	// Initialize connects and initializes the MCP server
	Initialize(ctx context.Context) error

	// ListTools returns all available tools
	ListTools(ctx context.Context) ([]transport.Tool, error)

	// CallTool executes a tool with the given arguments
	CallTool(ctx context.Context, name string, arguments map[string]interface{}) (*transport.ToolResponse, error)

	// Close closes the client connection
	Close() error

	// GetName returns the name of the MCP server
	GetName() string

	// GetPrefix returns the tool name prefix
	GetPrefix() string
}

// MCPClient implements the Client interface
type MCPClient struct {
	config      config.MCPConfig
	transport   transport.Transport
	mu          sync.RWMutex
	initialized bool
}

// NewClient creates a new MCP client based on configuration
func NewClient(cfg config.MCPConfig) (Client, error) {
	var t transport.Transport

	switch cfg.Transport {
	case "http", "":
		t = transport.NewHTTPTransport(cfg.URL)
		// Set auth headers if provided
		if cfg.Auth != nil {
			httpTransport := t.(*transport.HTTPTransport)
			for key, value := range cfg.Auth {
				httpTransport.SetHeader(key, value)
			}
		}
	default:
		return nil, fmt.Errorf("unsupported transport: %s", cfg.Transport)
	}

	return &MCPClient{
		config:    cfg,
		transport: t,
	}, nil
}

// Initialize connects and initializes the MCP server
func (c *MCPClient) Initialize(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.initialized {
		return nil
	}

	if err := c.transport.Initialize(ctx, nil); err != nil {
		return fmt.Errorf("failed to initialize client %s: %w", c.config.Name, err)
	}

	c.initialized = true
	return nil
}

// ensureInitialized ensures the client is initialized (lazy initialization)
func (c *MCPClient) ensureInitialized(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.initialized {
		return nil
	}

	if err := c.transport.Initialize(ctx, nil); err != nil {
		return fmt.Errorf("failed to initialize client %s: %w", c.config.Name, err)
	}

	c.initialized = true
	return nil
}

// ListTools returns all available tools
func (c *MCPClient) ListTools(ctx context.Context) ([]transport.Tool, error) {
	// Lazy initialization - initialize if not already done
	if err := c.ensureInitialized(ctx); err != nil {
		return nil, err
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	tools, err := c.transport.ListTools(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list tools from %s: %w", c.config.Name, err)
	}

	// Apply prefix to tool names if configured
	if c.config.Prefix != "" {
		for i := range tools {
			tools[i].Name = c.config.Prefix + tools[i].Name
		}
	}

	return tools, nil
}

// CallTool executes a tool with the given arguments
func (c *MCPClient) CallTool(ctx context.Context, name string, arguments map[string]interface{}) (*transport.ToolResponse, error) {
	// Lazy initialization - initialize if not already done
	if err := c.ensureInitialized(ctx); err != nil {
		return nil, err
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	// Remove prefix if present
	actualName := name
	if c.config.Prefix != "" && len(name) > len(c.config.Prefix) {
		if name[:len(c.config.Prefix)] == c.config.Prefix {
			actualName = name[len(c.config.Prefix):]
		}
	}

	resp, err := c.transport.CallTool(ctx, actualName, arguments)
	if err != nil {
		return nil, fmt.Errorf("failed to call tool %s on %s: %w", name, c.config.Name, err)
	}

	return resp, nil
}

// Close closes the client connection
func (c *MCPClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.transport != nil {
		return c.transport.Close()
	}
	return nil
}

// GetName returns the name of the MCP server
func (c *MCPClient) GetName() string {
	return c.config.Name
}

// GetPrefix returns the tool name prefix
func (c *MCPClient) GetPrefix() string {
	return c.config.Prefix
}
