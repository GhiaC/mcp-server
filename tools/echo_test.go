package tools

import (
	"testing"
)

func TestGetEchoTool(t *testing.T) {
	tool := GetEchoTool()

	if tool.Name != "echo" {
		t.Errorf("Expected tool name 'echo', got '%s'", tool.Name)
	}

	if tool.Description != "Echo back the provided message" {
		t.Errorf("Expected description 'Echo back the provided message', got '%s'", tool.Description)
	}

	// Verify input schema structure
	if tool.InputSchema == nil {
		t.Fatal("InputSchema should not be nil")
	}

	schemaType, ok := tool.InputSchema["type"].(string)
	if !ok || schemaType != "object" {
		t.Errorf("Expected schema type 'object', got %v", tool.InputSchema["type"])
	}

	properties, ok := tool.InputSchema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected properties to be a map")
	}

	messageProp, ok := properties["message"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected message property to be a map")
	}

	messageType, ok := messageProp["type"].(string)
	if !ok || messageType != "string" {
		t.Errorf("Expected message type 'string', got %v", messageProp["type"])
	}

	required, ok := tool.InputSchema["required"].([]string)
	if !ok {
		t.Fatal("Expected required to be a []string")
	}

	if len(required) != 1 || required[0] != "message" {
		t.Errorf("Expected required to contain 'message', got %v", required)
	}
}

func TestCallEcho(t *testing.T) {
	arguments := map[string]interface{}{
		"message": "Test message",
	}

	result, err := CallEcho(arguments)
	if err != nil {
		t.Fatalf("CallEcho returned error: %v", err)
	}

	if result != "Test message" {
		t.Errorf("Expected 'Test message', got '%s'", result)
	}
}

func TestCallEchoEmptyMessage(t *testing.T) {
	arguments := map[string]interface{}{
		"message": "",
	}

	result, err := CallEcho(arguments)
	if err != nil {
		t.Fatalf("CallEcho returned error: %v", err)
	}

	if result != "" {
		t.Errorf("Expected empty string, got '%s'", result)
	}
}

func TestCallEchoMissingMessage(t *testing.T) {
	arguments := map[string]interface{}{}

	_, err := CallEcho(arguments)
	if err == nil {
		t.Fatal("Expected error for missing message argument")
	}
}

func TestCallEchoInvalidType(t *testing.T) {
	arguments := map[string]interface{}{
		"message": 123, // Should be string
	}

	_, err := CallEcho(arguments)
	if err == nil {
		t.Fatal("Expected error for invalid message type")
	}
}

func TestCallEchoLongMessage(t *testing.T) {
	longMessage := "This is a very long message that contains many characters. " +
		"It should be echoed back correctly without any issues. " +
		"Testing the echo tool with a longer string to ensure it handles " +
		"messages of various lengths properly."

	arguments := map[string]interface{}{
		"message": longMessage,
	}

	result, err := CallEcho(arguments)
	if err != nil {
		t.Fatalf("CallEcho returned error: %v", err)
	}

	if result != longMessage {
		t.Errorf("Expected long message to be echoed correctly")
	}
}
