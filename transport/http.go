package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPTransport implements Transport interface using HTTP
type HTTPTransport struct {
	baseURL    string
	httpClient *http.Client
	headers    map[string]string
}

// NewHTTPTransport creates a new HTTP transport
func NewHTTPTransport(baseURL string) *HTTPTransport {
	return &HTTPTransport{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		headers: make(map[string]string),
	}
}

// SetHeader sets a custom header for all requests
func (t *HTTPTransport) SetHeader(key, value string) {
	t.headers[key] = value
}

// Initialize connects to the MCP server and initializes the connection
func (t *HTTPTransport) Initialize(ctx context.Context, config map[string]interface{}) error {
	req, err := http.NewRequestWithContext(ctx, "GET", t.baseURL+"/initialize", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	for k, v := range t.headers {
		req.Header.Set(k, v)
	}

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to initialize: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("initialize failed with status %d: %s", resp.StatusCode, string(body))
	}

	var initResp InitializeResponse
	if err := json.NewDecoder(resp.Body).Decode(&initResp); err != nil {
		return fmt.Errorf("failed to decode initialize response: %w", err)
	}

	// Validate protocol version
	if initResp.ProtocolVersion != "2024-11-05" {
		return fmt.Errorf("unsupported protocol version: %s", initResp.ProtocolVersion)
	}

	return nil
}

// ListTools returns all available tools from the remote MCP server
func (t *HTTPTransport) ListTools(ctx context.Context) ([]Tool, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", t.baseURL+"/tools/list", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	for k, v := range t.headers {
		req.Header.Set(k, v)
	}

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list tools: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list tools failed with status %d: %s", resp.StatusCode, string(body))
	}

	var toolsResp ToolsListResponse
	if err := json.NewDecoder(resp.Body).Decode(&toolsResp); err != nil {
		return nil, fmt.Errorf("failed to decode tools list response: %w", err)
	}

	return toolsResp.Tools, nil
}

// CallTool executes a tool on the remote MCP server
func (t *HTTPTransport) CallTool(ctx context.Context, name string, arguments map[string]interface{}) (*ToolResponse, error) {
	requestBody := map[string]interface{}{
		"name":      name,
		"arguments": arguments,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", t.baseURL+"/tools/call", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	for k, v := range t.headers {
		req.Header.Set(k, v)
	}

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call tool: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("tool '%s' not found", name)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("tool call failed with status %d: %s", resp.StatusCode, string(body))
	}

	var toolResp ToolResponse
	if err := json.NewDecoder(resp.Body).Decode(&toolResp); err != nil {
		return nil, fmt.Errorf("failed to decode tool response: %w", err)
	}

	return &toolResp, nil
}

// Close closes the transport connection (no-op for HTTP)
func (t *HTTPTransport) Close() error {
	return nil
}
