package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// HTTPTransport implements Transport interface using HTTP
type HTTPTransport struct {
	baseURL           string
	httpClient        *http.Client
	headers           map[string]string
	sessionID         string // Session ID for streamable-http (Cloudflare)
	useStreamableHTTP bool   // Whether to use streamable-http protocol
	requestID         int    // Counter for JSON-RPC request IDs
}

// NewHTTPTransport creates a new HTTP transport
func NewHTTPTransport(baseURL string) *HTTPTransport {
	// Detect if this is a Cloudflare MCP server (uses streamable-http)
	useStreamableHTTP := strings.Contains(baseURL, "mcp.cloudflare.com")

	return &HTTPTransport{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		headers:           make(map[string]string),
		useStreamableHTTP: useStreamableHTTP,
		requestID:         1,
	}
}

// SetHeader sets a custom header for all requests
func (t *HTTPTransport) SetHeader(key, value string) {
	t.headers[key] = value
}

// Initialize connects to the MCP server and initializes the connection
func (t *HTTPTransport) Initialize(ctx context.Context, config map[string]interface{}) error {
	if t.useStreamableHTTP {
		return t.initializeStreamableHTTP(ctx)
	}
	return t.initializeREST(ctx)
}

// initializeREST initializes using REST-style endpoints
func (t *HTTPTransport) initializeREST(ctx context.Context) error {
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

// initializeStreamableHTTP initializes using JSON-RPC 2.0 over streamable-http
func (t *HTTPTransport) initializeStreamableHTTP(ctx context.Context) error {
	// Create JSON-RPC 2.0 initialize request
	requestID := t.requestID
	t.requestID++

	jsonRPCRequest := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "initialize",
		"params": map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
			"clientInfo": map[string]interface{}{
				"name":    "mcp-go-client",
				"version": "1.0.0",
			},
		},
		"id": requestID,
	}

	bodyBytes, err := json.Marshal(jsonRPCRequest)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON-RPC request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", t.baseURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	for k, v := range t.headers {
		req.Header.Set(k, v)
	}

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to initialize: %w", err)
	}
	defer resp.Body.Close()

	// Extract session ID from response header
	if sessionID := resp.Header.Get("Mcp-Session-Id"); sessionID != "" {
		t.sessionID = sessionID
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("initialize failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse JSON-RPC response
	var jsonRPCResp struct {
		JSONRPC string `json:"jsonrpc"`
		Result  struct {
			ProtocolVersion string                 `json:"protocolVersion"`
			Capabilities    map[string]interface{} `json:"capabilities"`
			ServerInfo      ServerInfo             `json:"serverInfo"`
		} `json:"result"`
		Error *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
		ID interface{} `json:"id"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&jsonRPCResp); err != nil {
		return fmt.Errorf("failed to decode JSON-RPC response: %w", err)
	}

	if jsonRPCResp.Error != nil {
		return fmt.Errorf("JSON-RPC error: %d - %s", jsonRPCResp.Error.Code, jsonRPCResp.Error.Message)
	}

	// Validate protocol version
	if jsonRPCResp.Result.ProtocolVersion != "2024-11-05" {
		return fmt.Errorf("unsupported protocol version: %s", jsonRPCResp.Result.ProtocolVersion)
	}

	return nil
}

// ListTools returns all available tools from the remote MCP server
func (t *HTTPTransport) ListTools(ctx context.Context) ([]Tool, error) {
	if t.useStreamableHTTP {
		return t.listToolsStreamableHTTP(ctx)
	}
	return t.listToolsREST(ctx)
}

// listToolsREST lists tools using REST-style endpoint
func (t *HTTPTransport) listToolsREST(ctx context.Context) ([]Tool, error) {
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

// listToolsStreamableHTTP lists tools using JSON-RPC 2.0
func (t *HTTPTransport) listToolsStreamableHTTP(ctx context.Context) ([]Tool, error) {
	requestID := t.requestID
	t.requestID++

	jsonRPCRequest := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "tools/list",
		"params":  map[string]interface{}{},
		"id":      requestID,
	}

	bodyBytes, err := json.Marshal(jsonRPCRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON-RPC request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", t.baseURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if t.sessionID != "" {
		req.Header.Set("Mcp-Session-Id", t.sessionID)
	}
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

	// Parse JSON-RPC response
	var jsonRPCResp struct {
		JSONRPC string `json:"jsonrpc"`
		Result  struct {
			Tools []Tool `json:"tools"`
		} `json:"result"`
		Error *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
		ID interface{} `json:"id"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&jsonRPCResp); err != nil {
		return nil, fmt.Errorf("failed to decode JSON-RPC response: %w", err)
	}

	if jsonRPCResp.Error != nil {
		return nil, fmt.Errorf("JSON-RPC error: %d - %s", jsonRPCResp.Error.Code, jsonRPCResp.Error.Message)
	}

	return jsonRPCResp.Result.Tools, nil
}

// CallTool executes a tool on the remote MCP server
func (t *HTTPTransport) CallTool(ctx context.Context, name string, arguments map[string]interface{}) (*ToolResponse, error) {
	if t.useStreamableHTTP {
		return t.callToolStreamableHTTP(ctx, name, arguments)
	}
	return t.callToolREST(ctx, name, arguments)
}

// callToolREST calls a tool using REST-style endpoint
func (t *HTTPTransport) callToolREST(ctx context.Context, name string, arguments map[string]interface{}) (*ToolResponse, error) {
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

// callToolStreamableHTTP calls a tool using JSON-RPC 2.0
func (t *HTTPTransport) callToolStreamableHTTP(ctx context.Context, name string, arguments map[string]interface{}) (*ToolResponse, error) {
	requestID := t.requestID
	t.requestID++

	jsonRPCRequest := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name":      name,
			"arguments": arguments,
		},
		"id": requestID,
	}

	bodyBytes, err := json.Marshal(jsonRPCRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON-RPC request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", t.baseURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if t.sessionID != "" {
		req.Header.Set("Mcp-Session-Id", t.sessionID)
	}
	for k, v := range t.headers {
		req.Header.Set(k, v)
	}

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call tool: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("tool call failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse JSON-RPC response
	var jsonRPCResp struct {
		JSONRPC string `json:"jsonrpc"`
		Result  struct {
			Content []ContentItem `json:"content"`
		} `json:"result"`
		Error *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
		ID interface{} `json:"id"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&jsonRPCResp); err != nil {
		return nil, fmt.Errorf("failed to decode JSON-RPC response: %w", err)
	}

	if jsonRPCResp.Error != nil {
		if jsonRPCResp.Error.Code == -32000 {
			return nil, fmt.Errorf("tool '%s' not found", name)
		}
		return nil, fmt.Errorf("JSON-RPC error: %d - %s", jsonRPCResp.Error.Code, jsonRPCResp.Error.Message)
	}

	return &ToolResponse{
		Content: jsonRPCResp.Result.Content,
	}, nil
}

// Close closes the transport connection (no-op for HTTP)
func (t *HTTPTransport) Close() error {
	return nil
}
