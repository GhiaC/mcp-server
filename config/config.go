package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// MCPConfig represents configuration for an MCP server connection
type MCPConfig struct {
	Name      string            `json:"name"`
	URL       string            `json:"url"`
	Transport string            `json:"transport"` // "http", "sse", "stdio"
	Auth      map[string]string `json:"auth"`      // Auth headers/credentials
	Enabled   bool              `json:"enabled"`
	Prefix    string            `json:"prefix"` // Tool name prefix (e.g., "cloudflare:")
}

// GooglePSEConfig represents Google PSE configuration
type GooglePSEConfig struct {
	APIKey         string `json:"api_key"`
	SearchEngineID string `json:"search_engine_id"`
	Enabled        bool   `json:"enabled"`
}

// Config represents the application configuration
type Config struct {
	Port        string          `json:"port"`         // Server port (default: ":3333")
	BearerToken string          `json:"bearer_token"` // Bearer token for authentication (optional)
	GooglePSE   GooglePSEConfig `json:"google_pse"`   // Google PSE configuration
	Servers     []MCPConfig     `json:"servers"`      // Remote MCP servers
}

// LoadConfig loads configuration from a JSON file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &config, nil
}

// LoadConfigFromEnv loads configuration from environment variables
// Format: MCP_SERVERS='[{"name":"cloudflare","url":"...","transport":"http"}]'
func LoadConfigFromEnv() (*Config, error) {
	serversJSON := os.Getenv("MCP_SERVERS")
	bearerToken := os.Getenv("MCP_BEARER_TOKEN")

	config := &Config{
		BearerToken: bearerToken,
		Servers:     []MCPConfig{},
	}

	if serversJSON == "" {
		return config, nil
	}

	var servers []MCPConfig
	if err := json.Unmarshal([]byte(serversJSON), &servers); err != nil {
		return nil, fmt.Errorf("failed to parse MCP_SERVERS env var: %w", err)
	}

	config.Servers = servers
	return config, nil
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Port:        ":3333",
		BearerToken: "",
		GooglePSE: GooglePSEConfig{
			Enabled: false,
		},
		Servers: []MCPConfig{},
	}
}

// GetPort returns the server port, defaulting to ":3333" if not set
func (c *Config) GetPort() string {
	if c.Port == "" {
		return ":3333"
	}
	// Ensure port starts with ":"
	if c.Port[0] != ':' {
		return ":" + c.Port
	}
	return c.Port
}

// GetGooglePSEConfig returns Google PSE configuration
func (c *Config) GetGooglePSEConfig() *GooglePSEConfig {
	return &c.GooglePSE
}

// GetBearerToken returns the bearer token for authentication
func (c *Config) GetBearerToken() string {
	return c.BearerToken
}
