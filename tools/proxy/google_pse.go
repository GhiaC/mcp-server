package proxy

import (
	"context"
	"fmt"
	"mcp-go/gateway"
	"mcp-go/transport"
)

// GooglePSEProxy provides a wrapper for Google PSE MCP tools
type GooglePSEProxy struct {
	gateway *gateway.Gateway
}

// NewGooglePSEProxy creates a new Google PSE proxy
func NewGooglePSEProxy(gw *gateway.Gateway) *GooglePSEProxy {
	return &GooglePSEProxy{
		gateway: gw,
	}
}

// Search performs a web search using Google PSE
func (p *GooglePSEProxy) Search(ctx context.Context, query string, num int) (*transport.ToolResponse, error) {
	if p.gateway == nil {
		return nil, fmt.Errorf("gateway not initialized")
	}

	arguments := map[string]interface{}{
		"query": query,
		"num":   num,
	}

	return p.gateway.CallTool(ctx, "google_pse:search", arguments)
}

// CallGooglePSETool calls a Google PSE tool by name
func (p *GooglePSEProxy) CallGooglePSETool(ctx context.Context, toolName string, arguments map[string]interface{}) (*transport.ToolResponse, error) {
	if p.gateway == nil {
		return nil, fmt.Errorf("gateway not initialized")
	}

	// Add prefix if not present
	fullToolName := toolName
	if len(toolName) < 11 || toolName[:11] != "google_pse:" {
		fullToolName = "google_pse:" + toolName
	}

	return p.gateway.CallTool(ctx, fullToolName, arguments)
}

// ListGooglePSETools lists all available Google PSE tools
func (p *GooglePSEProxy) ListGooglePSETools(ctx context.Context) ([]transport.Tool, error) {
	if p.gateway == nil {
		return nil, fmt.Errorf("gateway not initialized")
	}

	allTools, err := p.gateway.ListAllTools(ctx)
	if err != nil {
		return nil, err
	}

	// Filter only Google PSE tools
	var googlePSETools []transport.Tool
	for _, tool := range allTools {
		if len(tool.Name) >= 11 && tool.Name[:11] == "google_pse:" {
			googlePSETools = append(googlePSETools, tool)
		}
	}

	return googlePSETools, nil
}
