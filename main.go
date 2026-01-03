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

	// Configure Google PSE from config file or environment variables
	googlePSE := cfg.GetGooglePSEConfig()
	var apiKey, searchEngineID string
	var googlePSEEnabled bool

	if googlePSE.Enabled && googlePSE.APIKey != "" && googlePSE.SearchEngineID != "" {
		// Use config file values
		apiKey = googlePSE.APIKey
		searchEngineID = googlePSE.SearchEngineID
		googlePSEEnabled = true
		log.Println("Google PSE configured from config file")
	} else {
		// Try environment variables
		apiKey = os.Getenv("GOOGLE_PSE_API_KEY")
		searchEngineID = os.Getenv("GOOGLE_PSE_SEARCH_ENGINE_ID")
		if apiKey != "" && searchEngineID != "" {
			googlePSEEnabled = true
			log.Println("Google PSE configured from environment variables")
		}
	}

	if googlePSEEnabled {
		tools.SetGooglePSEConfig(apiKey, searchEngineID)
		log.Println("Google PSE enabled successfully")
	} else {
		log.Println("Google PSE not configured (set enabled:true in config file or GOOGLE_PSE_API_KEY and GOOGLE_PSE_SEARCH_ENGINE_ID env vars)")
	}

	// Get bearer token from config or environment
	bearerToken := cfg.GetBearerToken()
	if bearerToken == "" {
		bearerToken = os.Getenv("MCP_BEARER_TOKEN")
	}

	// Start server with gateway, configured port, and bearer token
	port := cfg.GetPort()
	server.StartWithGatewayAndPortAndAuth(gw, port, bearerToken)
}
