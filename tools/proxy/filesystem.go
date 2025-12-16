package proxy

import (
	"context"
	"fmt"
	"mcp-go/gateway"
	"mcp-go/transport"
)

// FileSystemProxy provides a wrapper for File System MCP tools
type FileSystemProxy struct {
	gateway *gateway.Gateway
}

// NewFileSystemProxy creates a new File System proxy
func NewFileSystemProxy(gw *gateway.Gateway) *FileSystemProxy {
	return &FileSystemProxy{
		gateway: gw,
	}
}

// ReadFile reads a file through the File System MCP
func (p *FileSystemProxy) ReadFile(ctx context.Context, filePath string) (*transport.ToolResponse, error) {
	if p.gateway == nil {
		return nil, fmt.Errorf("gateway not initialized")
	}

	arguments := map[string]interface{}{
		"path": filePath,
	}

	return p.gateway.CallTool(ctx, "filesystem:read_file", arguments)
}

// WriteFile writes content to a file through the File System MCP
func (p *FileSystemProxy) WriteFile(ctx context.Context, filePath string, content string) (*transport.ToolResponse, error) {
	if p.gateway == nil {
		return nil, fmt.Errorf("gateway not initialized")
	}

	arguments := map[string]interface{}{
		"path":    filePath,
		"content": content,
	}

	return p.gateway.CallTool(ctx, "filesystem:write_file", arguments)
}

// ListDirectory lists files in a directory through the File System MCP
func (p *FileSystemProxy) ListDirectory(ctx context.Context, dirPath string) (*transport.ToolResponse, error) {
	if p.gateway == nil {
		return nil, fmt.Errorf("gateway not initialized")
	}

	arguments := map[string]interface{}{
		"path": dirPath,
	}

	return p.gateway.CallTool(ctx, "filesystem:list_directory", arguments)
}

// CreateDirectory creates a directory through the File System MCP
func (p *FileSystemProxy) CreateDirectory(ctx context.Context, dirPath string) (*transport.ToolResponse, error) {
	if p.gateway == nil {
		return nil, fmt.Errorf("gateway not initialized")
	}

	arguments := map[string]interface{}{
		"path": dirPath,
	}

	return p.gateway.CallTool(ctx, "filesystem:create_directory", arguments)
}

// DeleteFile deletes a file through the File System MCP
func (p *FileSystemProxy) DeleteFile(ctx context.Context, filePath string) (*transport.ToolResponse, error) {
	if p.gateway == nil {
		return nil, fmt.Errorf("gateway not initialized")
	}

	arguments := map[string]interface{}{
		"path": filePath,
	}

	return p.gateway.CallTool(ctx, "filesystem:delete_file", arguments)
}

// ListFileSystemTools lists all available File System tools
func (p *FileSystemProxy) ListFileSystemTools(ctx context.Context) ([]transport.Tool, error) {
	if p.gateway == nil {
		return nil, fmt.Errorf("gateway not initialized")
	}

	allTools, err := p.gateway.ListAllTools(ctx)
	if err != nil {
		return nil, err
	}

	// Filter only File System tools
	var filesystemTools []transport.Tool
	for _, tool := range allTools {
		if len(tool.Name) >= 11 && tool.Name[:11] == "filesystem:" {
			filesystemTools = append(filesystemTools, tool)
		}
	}

	return filesystemTools, nil
}

// CallFileSystemTool calls a File System tool by name
func (p *FileSystemProxy) CallFileSystemTool(ctx context.Context, toolName string, arguments map[string]interface{}) (*transport.ToolResponse, error) {
	if p.gateway == nil {
		return nil, fmt.Errorf("gateway not initialized")
	}

	// Add prefix if not present
	fullToolName := toolName
	if len(toolName) < 11 || toolName[:11] != "filesystem:" {
		fullToolName = "filesystem:" + toolName
	}

	return p.gateway.CallTool(ctx, fullToolName, arguments)
}
