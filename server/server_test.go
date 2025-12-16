package server

import (
	"bytes"
	"encoding/json"
	"mcp-go/tools"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleInitialize(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/initialize", nil)
	w := httptest.NewRecorder()

	handleInitialize(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response InitializeResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.ProtocolVersion != "2024-11-05" {
		t.Errorf("Expected protocol version '2024-11-05', got '%s'", response.ProtocolVersion)
	}

	if tools, ok := response.Capabilities["tools"].(bool); !ok || !tools {
		t.Errorf("Expected tools capability to be true, got %v", response.Capabilities["tools"])
	}

	if response.ServerInfo.Name != "mcp-go" {
		t.Errorf("Expected server name 'mcp-go', got '%s'", response.ServerInfo.Name)
	}

	if response.ServerInfo.Version != "0.1.0" {
		t.Errorf("Expected server version '0.1.0', got '%s'", response.ServerInfo.Version)
	}
}

func TestHandleInitializeMethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/initialize", nil)
	w := httptest.NewRecorder()

	handleInitialize(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status code %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestHandleToolsList(t *testing.T) {
	srv := NewServer(nil) // No gateway for this test
	req := httptest.NewRequest(http.MethodGet, "/tools/list", nil)
	w := httptest.NewRecorder()

	srv.handleToolsList(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response ToolsListResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(response.Tools) != 1 {
		t.Fatalf("Expected 1 tool, got %d", len(response.Tools))
	}

	// Type assert to EchoTool
	toolMap, ok := response.Tools[0].(map[string]interface{})
	if !ok {
		// Try to unmarshal as EchoTool
		toolJSON, _ := json.Marshal(response.Tools[0])
		var echoTool tools.EchoTool
		if err := json.Unmarshal(toolJSON, &echoTool); err == nil {
			if echoTool.Name != "echo" {
				t.Errorf("Expected tool name 'echo', got '%s'", echoTool.Name)
			}
			return
		}
		t.Fatalf("Failed to parse tool")
	}

	if toolMap["name"] != "echo" {
		t.Errorf("Expected tool name 'echo', got '%v'", toolMap["name"])
	}

	if toolMap["description"] != "Echo back the provided message" {
		t.Errorf("Expected description 'Echo back the provided message', got '%v'", toolMap["description"])
	}

	// Verify input schema
	if inputSchema, ok := toolMap["inputSchema"].(map[string]interface{}); ok {
		if schemaType, ok := inputSchema["type"].(string); !ok || schemaType != "object" {
			t.Errorf("Expected input schema type 'object', got %v", inputSchema["type"])
		}
	}
}

func TestHandleToolsListMethodNotAllowed(t *testing.T) {
	srv := NewServer(nil)
	req := httptest.NewRequest(http.MethodPost, "/tools/list", nil)
	w := httptest.NewRecorder()

	srv.handleToolsList(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status code %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestHandleToolsCallEcho(t *testing.T) {
	srv := NewServer(nil)
	requestBody := ToolCallRequest{
		Name: "echo",
		Arguments: map[string]interface{}{
			"message": "Hello, World!",
		},
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/tools/call", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.handleToolsCall(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response ToolCallResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(response.Content) != 1 {
		t.Fatalf("Expected 1 content item, got %d", len(response.Content))
	}

	content := response.Content[0]
	if content.Type != "text" {
		t.Errorf("Expected content type 'text', got '%s'", content.Type)
	}

	if content.Text != "Hello, World!" {
		t.Errorf("Expected text 'Hello, World!', got '%s'", content.Text)
	}
}

func TestHandleToolsCallUnknownTool(t *testing.T) {
	srv := NewServer(nil)
	requestBody := ToolCallRequest{
		Name:      "unknown-tool",
		Arguments: map[string]interface{}{},
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/tools/call", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.handleToolsCall(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestHandleToolsCallInvalidJSON(t *testing.T) {
	srv := NewServer(nil)
	req := httptest.NewRequest(http.MethodPost, "/tools/call", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.handleToolsCall(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHandleToolsCallMethodNotAllowed(t *testing.T) {
	srv := NewServer(nil)
	req := httptest.NewRequest(http.MethodGet, "/tools/call", nil)
	w := httptest.NewRecorder()

	srv.handleToolsCall(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status code %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}
