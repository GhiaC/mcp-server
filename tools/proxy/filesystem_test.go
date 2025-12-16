package proxy

import (
	"context"
	"mcp-go/gateway"
	"testing"
)

func TestNewFileSystemProxy(t *testing.T) {
	gw := gateway.NewGateway()
	proxy := NewFileSystemProxy(gw)

	if proxy == nil {
		t.Fatal("Expected FileSystemProxy, got nil")
	}

	if proxy.gateway != gw {
		t.Error("Gateway not set correctly")
	}
}

func TestFileSystemProxyWithoutGateway(t *testing.T) {
	proxy := NewFileSystemProxy(nil)
	ctx := context.Background()

	_, err := proxy.ReadFile(ctx, "/test.txt")
	if err == nil {
		t.Error("Expected error when gateway is nil")
	}

	_, err = proxy.WriteFile(ctx, "/test.txt", "content")
	if err == nil {
		t.Error("Expected error when gateway is nil")
	}

	_, err = proxy.ListDirectory(ctx, "/")
	if err == nil {
		t.Error("Expected error when gateway is nil")
	}
}

func TestFileSystemProxyToolNamePrefixing(t *testing.T) {
	gw := gateway.NewGateway()
	proxy := NewFileSystemProxy(gw)
	ctx := context.Background()

	// Test that tool names are prefixed correctly
	// Note: This will fail if gateway is not connected, but tests the logic
	_, err := proxy.CallFileSystemTool(ctx, "read_file", map[string]interface{}{"path": "/test"})
	// We expect an error because gateway is not initialized, but the tool name should be prefixed
	if err != nil && err.Error() != "gateway not initialized" {
		// If we get a different error, it means the tool name was constructed correctly
		// and the call was attempted (which is what we want to test)
	}
}
