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

// Config represents the application configuration
type Config struct {
	Servers []MCPConfig `json:"servers"`
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
	if serversJSON == "" {
		return &Config{Servers: []MCPConfig{}}, nil
	}

	var servers []MCPConfig
	if err := json.Unmarshal([]byte(serversJSON), &servers); err != nil {
		return nil, fmt.Errorf("failed to parse MCP_SERVERS env var: %w", err)
	}

	return &Config{Servers: servers}, nil
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Servers: []MCPConfig{},
	}
}
