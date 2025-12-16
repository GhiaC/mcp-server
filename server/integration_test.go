package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestFullWorkflow tests the complete MCP workflow:
// 1. Initialize
// 2. List tools
// 3. Call echo tool
func TestFullWorkflow(t *testing.T) {
	// Step 1: Initialize
	initReq := httptest.NewRequest(http.MethodGet, "/initialize", nil)
	initW := httptest.NewRecorder()
	handleInitialize(initW, initReq)

	if initW.Code != http.StatusOK {
		t.Fatalf("Initialize failed with status %d", initW.Code)
	}

	var initResponse InitializeResponse
	if err := json.Unmarshal(initW.Body.Bytes(), &initResponse); err != nil {
		t.Fatalf("Failed to unmarshal initialize response: %v", err)
	}

	if initResponse.ProtocolVersion != "2024-11-05" {
		t.Errorf("Invalid protocol version: %s", initResponse.ProtocolVersion)
	}

	// Step 2: List tools
	srv := NewServer(nil)
	listReq := httptest.NewRequest(http.MethodGet, "/tools/list", nil)
	listW := httptest.NewRecorder()
	srv.handleToolsList(listW, listReq)

	if listW.Code != http.StatusOK {
		t.Fatalf("List tools failed with status %d", listW.Code)
	}

	var listResponse ToolsListResponse
	if err := json.Unmarshal(listW.Body.Bytes(), &listResponse); err != nil {
		t.Fatalf("Failed to unmarshal list response: %v", err)
	}

	if len(listResponse.Tools) == 0 {
		t.Fatal("No tools returned")
	}

	// Step 3: Call echo tool
	callBody := ToolCallRequest{
		Name: "echo",
		Arguments: map[string]interface{}{
			"message": "Integration test message",
		},
	}

	bodyBytes, _ := json.Marshal(callBody)
	callReq := httptest.NewRequest(http.MethodPost, "/tools/call", bytes.NewBuffer(bodyBytes))
	callReq.Header.Set("Content-Type", "application/json")
	callW := httptest.NewRecorder()
	srv.handleToolsCall(callW, callReq)

	if callW.Code != http.StatusOK {
		t.Fatalf("Call tool failed with status %d", callW.Code)
	}

	var callResponse ToolCallResponse
	if err := json.Unmarshal(callW.Body.Bytes(), &callResponse); err != nil {
		t.Fatalf("Failed to unmarshal call response: %v", err)
	}

	if len(callResponse.Content) != 1 {
		t.Fatalf("Expected 1 content item, got %d", len(callResponse.Content))
	}

	if callResponse.Content[0].Text != "Integration test message" {
		t.Errorf("Expected 'Integration test message', got '%s'", callResponse.Content[0].Text)
	}
}

// TestMCPResponseFormat verifies that tool call responses follow MCP format
func TestMCPResponseFormat(t *testing.T) {
	callBody := ToolCallRequest{
		Name: "echo",
		Arguments: map[string]interface{}{
			"message": "Format test",
		},
	}

	srv := NewServer(nil)
	bodyBytes, _ := json.Marshal(callBody)
	req := httptest.NewRequest(http.MethodPost, "/tools/call", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.handleToolsCall(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var response ToolCallResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify MCP format: content must be an array
	if response.Content == nil {
		t.Fatal("Content must not be nil")
	}

	if len(response.Content) == 0 {
		t.Fatal("Content array must not be empty")
	}

	// Verify content item structure
	item := response.Content[0]
	if item.Type != "text" {
		t.Errorf("Content type must be 'text', got '%s'", item.Type)
	}

	if item.Text == "" {
		t.Error("Content text must not be empty")
	}
}
