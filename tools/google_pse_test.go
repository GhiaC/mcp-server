package tools

import (
	"os"
	"testing"
)

func TestGetGooglePSETool(t *testing.T) {
	tool := GetGooglePSETool()

	if tool.Name != "google_pse_search" {
		t.Errorf("Expected tool name 'google_pse_search', got '%s'", tool.Name)
	}

	if tool.Description == "" {
		t.Error("Expected non-empty description")
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

	queryProp, ok := properties["query"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected query property to be a map")
	}

	queryType, ok := queryProp["type"].(string)
	if !ok || queryType != "string" {
		t.Errorf("Expected query type 'string', got %v", queryProp["type"])
	}

	required, ok := tool.InputSchema["required"].([]string)
	if !ok {
		t.Fatal("Expected required to be a []string")
	}

	if len(required) != 1 || required[0] != "query" {
		t.Errorf("Expected required to contain 'query', got %v", required)
	}
}

func TestSetGooglePSEConfig(t *testing.T) {
	SetGooglePSEConfig("test-api-key", "test-search-engine-id")

	config := GetGooglePSEConfig()
	if config == nil {
		t.Fatal("Expected config to be set")
	}

	if config.APIKey != "test-api-key" {
		t.Errorf("Expected API key 'test-api-key', got '%s'", config.APIKey)
	}

	if config.SearchEngineID != "test-search-engine-id" {
		t.Errorf("Expected Search Engine ID 'test-search-engine-id', got '%s'", config.SearchEngineID)
	}
}

func TestCallGooglePSEMissingConfig(t *testing.T) {
	// Reset config
	googlePSEConfig = nil

	arguments := map[string]interface{}{
		"query": "test query",
	}

	_, err := CallGooglePSE(arguments)
	if err == nil {
		t.Fatal("Expected error when config is not set")
	}

	if err.Error() != "Google PSE not configured. Please set API key and Search Engine ID" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestCallGooglePSEMissingQuery(t *testing.T) {
	SetGooglePSEConfig("test-key", "test-id")

	arguments := map[string]interface{}{}

	_, err := CallGooglePSE(arguments)
	if err == nil {
		t.Fatal("Expected error when query is missing")
	}
}

func TestCallGooglePSEEmptyQuery(t *testing.T) {
	SetGooglePSEConfig("test-key", "test-id")

	arguments := map[string]interface{}{
		"query": "",
	}

	_, err := CallGooglePSE(arguments)
	if err == nil {
		t.Fatal("Expected error when query is empty")
	}
}

func TestCallGooglePSEWithNum(t *testing.T) {
	SetGooglePSEConfig("test-key", "test-id")

	arguments := map[string]interface{}{
		"query": "test",
		"num":   float64(5),
	}

	// This will fail because we don't have real API credentials in test
	// But we can test that the function accepts the parameters
	_, err := CallGooglePSE(arguments)
	// We expect an error from the API call, not from parameter validation
	if err != nil && err.Error() == "query argument is required and must be a non-empty string" {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestGooglePSEConfigFromEnv(t *testing.T) {
	// Test that config can be set from environment
	os.Setenv("GOOGLE_PSE_API_KEY", "env-test-key")
	os.Setenv("GOOGLE_PSE_SEARCH_ENGINE_ID", "env-test-id")

	// Note: This is just testing the concept, actual usage is in main.go
	apiKey := os.Getenv("GOOGLE_PSE_API_KEY")
	searchEngineID := os.Getenv("GOOGLE_PSE_SEARCH_ENGINE_ID")

	if apiKey != "env-test-key" {
		t.Errorf("Expected 'env-test-key', got '%s'", apiKey)
	}

	if searchEngineID != "env-test-id" {
		t.Errorf("Expected 'env-test-id', got '%s'", searchEngineID)
	}

	os.Unsetenv("GOOGLE_PSE_API_KEY")
	os.Unsetenv("GOOGLE_PSE_SEARCH_ENGINE_ID")
}
