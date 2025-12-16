package main

import (
	"context"
	"log"
	"mcp-go/config"
	"mcp-go/gateway"
	"mcp-go/server"
	"mcp-go/tools"
	"os"
)

func main() {
	// Create gateway
	gw := gateway.NewGateway()

	// Try to load configuration from file or environment
	cfg, err := config.LoadConfig("mcp-config.json")
	if err != nil {
		// Try environment variables
		cfg, err = config.LoadConfigFromEnv()
		if err != nil {
			log.Printf("No configuration found, running without remote MCP servers: %v", err)
			cfg = config.DefaultConfig()
		}
	}

	// Load clients from configuration
	if err := gw.LoadFromConfig(cfg); err != nil {
		log.Fatalf("Failed to load MCP clients: %v", err)
	}

	// Initialize all clients
	ctx := context.Background()
	if err := gw.InitializeAll(ctx); err != nil {
		log.Printf("Warning: Some MCP clients failed to initialize: %v", err)
		log.Println("Server will continue with available clients")
	}

	// Configure Google PSE if environment variables are set
	apiKey := os.Getenv("GOOGLE_PSE_API_KEY")
	searchEngineID := os.Getenv("GOOGLE_PSE_SEARCH_ENGINE_ID")
	if apiKey != "" && searchEngineID != "" {
		tools.SetGooglePSEConfig(apiKey, searchEngineID)
		log.Println("Google PSE configured successfully")
	} else {
		log.Println("Google PSE not configured (set GOOGLE_PSE_API_KEY and GOOGLE_PSE_SEARCH_ENGINE_ID env vars)")
	}

	// Start server with gateway
	server.StartWithGateway(gw)
}
