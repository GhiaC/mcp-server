package gateway

import (
	"context"
	"fmt"
	"log"
	"mcp-go/client"
	"mcp-go/config"
	"mcp-go/transport"
	"strings"
	"sync"
)

// Gateway manages multiple MCP client connections
type Gateway struct {
	clients map[string]client.Client
	mu      sync.RWMutex
}

// NewGateway creates a new gateway instance
func NewGateway() *Gateway {
	return &Gateway{
		clients: make(map[string]client.Client),
	}
}

// AddClient adds a new MCP client to the gateway
func (g *Gateway) AddClient(c client.Client) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	name := c.GetName()
	if _, exists := g.clients[name]; exists {
		return fmt.Errorf("client %s already exists", name)
	}

	g.clients[name] = c
	return nil
}

// InitializeAll initializes all registered clients
func (g *Gateway) InitializeAll(ctx context.Context) error {
	g.mu.RLock()
	clients := make([]client.Client, 0, len(g.clients))
	for _, c := range g.clients {
		clients = append(clients, c)
	}
	g.mu.RUnlock()

	var errors []string
	for _, c := range clients {
		if err := c.Initialize(ctx); err != nil {
			log.Printf("Warning: Failed to initialize client %s: %v", c.GetName(), err)
			errors = append(errors, fmt.Sprintf("%s: %v", c.GetName(), err))
		} else {
			log.Printf("Successfully initialized MCP client: %s", c.GetName())
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("some clients failed to initialize: %s", strings.Join(errors, "; "))
	}

	return nil
}

// ListAllTools returns all tools from all connected clients
// Tools are fetched in parallel for better performance
func (g *Gateway) ListAllTools(ctx context.Context) ([]transport.Tool, error) {
	g.mu.RLock()
	clients := make([]client.Client, 0, len(g.clients))
	for _, c := range g.clients {
		clients = append(clients, c)
	}
	g.mu.RUnlock()

	// Use a channel to collect results from parallel goroutines
	type result struct {
		tools []transport.Tool
		err   error
		name  string
	}
	results := make(chan result, len(clients))

	// Fetch tools from all clients in parallel
	for _, c := range clients {
		go func(client client.Client) {
			tools, err := client.ListTools(ctx)
			results <- result{tools: tools, err: err, name: client.GetName()}
		}(c)
	}

	// Collect results
	var allTools []transport.Tool
	for i := 0; i < len(clients); i++ {
		res := <-results
		if res.err != nil {
			log.Printf("Warning: Failed to list tools from %s: %v", res.name, res.err)
			continue
		}
		allTools = append(allTools, res.tools...)
	}

	return allTools, nil
}

// CallTool calls a tool, routing to the appropriate client
func (g *Gateway) CallTool(ctx context.Context, name string, arguments map[string]interface{}) (*transport.ToolResponse, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	// Try to find the client that owns this tool
	for _, c := range g.clients {
		prefix := c.GetPrefix()
		if prefix != "" && strings.HasPrefix(name, prefix) {
			return c.CallTool(ctx, name, arguments)
		}
	}

	// If no prefix match, try all clients (for tools without prefix)
	for _, c := range g.clients {
		resp, err := c.CallTool(ctx, name, arguments)
		if err == nil {
			return resp, nil
		}
		// Continue to next client if tool not found
		if !strings.Contains(err.Error(), "not found") {
			return nil, err
		}
	}

	return nil, fmt.Errorf("tool '%s' not found in any connected MCP server", name)
}

// GetClient returns a client by name
func (g *Gateway) GetClient(name string) (client.Client, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	c, ok := g.clients[name]
	return c, ok
}

// CloseAll closes all client connections
func (g *Gateway) CloseAll() error {
	g.mu.Lock()
	defer g.mu.Unlock()

	var errors []string
	for name, c := range g.clients {
		if err := c.Close(); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", name, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors closing clients: %s", strings.Join(errors, "; "))
	}

	return nil
}

// LoadFromConfig loads clients from configuration
func (g *Gateway) LoadFromConfig(cfg *config.Config) error {
	for _, serverCfg := range cfg.Servers {
		if !serverCfg.Enabled {
			log.Printf("Skipping disabled MCP server: %s", serverCfg.Name)
			continue
		}

		c, err := client.NewClient(serverCfg)
		if err != nil {
			return fmt.Errorf("failed to create client for %s: %w", serverCfg.Name, err)
		}

		if err := g.AddClient(c); err != nil {
			return fmt.Errorf("failed to add client %s: %w", serverCfg.Name, err)
		}
	}

	return nil
}
