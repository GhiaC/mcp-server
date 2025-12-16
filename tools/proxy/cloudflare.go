package proxy

import (
	"context"
	"fmt"
	"mcp-go/gateway"
	"mcp-go/transport"
)

// CloudflareProxy provides a wrapper for Cloudflare MCP tools
type CloudflareProxy struct {
	gateway *gateway.Gateway
}

// NewCloudflareProxy creates a new Cloudflare proxy
func NewCloudflareProxy(gw *gateway.Gateway) *CloudflareProxy {
	return &CloudflareProxy{
		gateway: gw,
	}
}

// CallCloudflareTool calls a Cloudflare tool through the gateway
func (p *CloudflareProxy) CallCloudflareTool(ctx context.Context, toolName string, arguments map[string]interface{}) (*transport.ToolResponse, error) {
	if p.gateway == nil {
		return nil, fmt.Errorf("gateway not initialized")
	}

	// Add prefix if not present
	fullToolName := toolName
	if len(toolName) < 11 || toolName[:11] != "cloudflare:" {
		fullToolName = "cloudflare:" + toolName
	}

	return p.gateway.CallTool(ctx, fullToolName, arguments)
}

// ListCloudflareTools lists all available Cloudflare tools
func (p *CloudflareProxy) ListCloudflareTools(ctx context.Context) ([]transport.Tool, error) {
	if p.gateway == nil {
		return nil, fmt.Errorf("gateway not initialized")
	}

	allTools, err := p.gateway.ListAllTools(ctx)
	if err != nil {
		return nil, err
	}

	// Filter only Cloudflare tools
	var cloudflareTools []transport.Tool
	for _, tool := range allTools {
		if len(tool.Name) >= 11 && tool.Name[:11] == "cloudflare:" {
			cloudflareTools = append(cloudflareTools, tool)
		}
	}

	return cloudflareTools, nil
}
