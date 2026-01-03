package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mcp-go/gateway"
	"mcp-go/tools"
	"net/http"
	"strings"
	"sync"
	"time"
)

// JSONRPCRequest represents a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string                 `json:"jsonrpc"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params"`
	ID      interface{}            `json:"id"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
	ID      interface{} `json:"id"`
}

// RPCError represents a JSON-RPC error
type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// InitializeResult represents the result of initialize method
type InitializeResult struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    map[string]interface{} `json:"capabilities"`
	ServerInfo      ServerInfo             `json:"serverInfo"`
}

// ServerInfo contains server information
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ToolsListResult represents the result of tools/list method
type ToolsListResult struct {
	Tools []interface{} `json:"tools"`
}

// ToolCallResult represents the result of tools/call method
type ToolCallResult struct {
	Content []ContentItem `json:"content"`
}

// ContentItem represents a content item in the tool call response
type ContentItem struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Session represents a client session
type Session struct {
	ID        string
	CreatedAt time.Time
}

// Server holds the server state including gateway and sessions
type Server struct {
	gateway  *gateway.Gateway
	sessions map[string]*Session
	mu       sync.RWMutex
}

// NewServer creates a new server instance
func NewServer(gw *gateway.Gateway) *Server {
	return &Server{
		gateway:  gw,
		sessions: make(map[string]*Session),
	}
}

// generateSessionID generates a unique session ID
func (s *Server) generateSessionID() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("session-%d", timestamp)
}

// getOrCreateSession gets an existing session or creates a new one
func (s *Server) getOrCreateSession(sessionID string) *Session {
	s.mu.Lock()
	defer s.mu.Unlock()

	if sessionID != "" {
		if session, exists := s.sessions[sessionID]; exists {
			return session
		}
	}

	// Create new session
	newSessionID := s.generateSessionID()
	session := &Session{
		ID:        newSessionID,
		CreatedAt: time.Now(),
	}
	s.sessions[newSessionID] = session
	return session
}

// writeSSEResponse writes a JSON-RPC response as SSE
func writeSSEResponse(w http.ResponseWriter, response JSONRPCResponse) error {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Mcp-Session-Id")

	// Marshal response to JSON
	jsonData, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	// Write SSE format: "data: {json}\n\n"
	_, err = fmt.Fprintf(w, "data: %s\n\n", string(jsonData))
	if err != nil {
		return fmt.Errorf("failed to write SSE response: %w", err)
	}

	// Flush the response
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}

	return nil
}

// writeJSONResponse writes a JSON-RPC response as regular JSON (fallback)
func writeJSONResponse(w http.ResponseWriter, response JSONRPCResponse) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Mcp-Session-Id")

	return json.NewEncoder(w).Encode(response)
}

// handleMCP handles the main MCP endpoint (POST /mcp)
func (s *Server) handleMCP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get or create session
	sessionID := r.Header.Get("Mcp-Session-Id")
	session := s.getOrCreateSession(sessionID)

	// Set session ID in response header
	w.Header().Set("Mcp-Session-Id", session.ID)

	// Parse JSON-RPC request
	var req JSONRPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResp := JSONRPCResponse{
			JSONRPC: "2.0",
			Error: &RPCError{
				Code:    -32700,
				Message: "Parse error",
			},
			ID: nil,
		}
		writeJSONResponse(w, errorResp)
		return
	}

	// Validate JSON-RPC version
	if req.JSONRPC != "2.0" {
		errorResp := JSONRPCResponse{
			JSONRPC: "2.0",
			Error: &RPCError{
				Code:    -32600,
				Message: "Invalid Request",
			},
			ID: req.ID,
		}
		writeJSONResponse(w, errorResp)
		return
	}

	// Check Accept header to determine response format
	// MCP streamable-http uses SSE by default, but allows JSON fallback
	acceptHeader := r.Header.Get("Accept")
	useSSE := acceptHeader == "" ||
		acceptHeader == "text/event-stream" ||
		acceptHeader == "*/*" ||
		(acceptHeader != "" && !strings.Contains(acceptHeader, "application/json"))

	// Route to appropriate handler
	var response JSONRPCResponse
	var err error

	switch req.Method {
	case "initialize":
		response, err = s.handleInitialize(req)
	case "tools/list":
		response, err = s.handleToolsList(r.Context(), req)
	case "tools/call":
		response, err = s.handleToolsCall(r.Context(), req)
	default:
		response = JSONRPCResponse{
			JSONRPC: "2.0",
			Error: &RPCError{
				Code:    -32601,
				Message: "Method not found",
			},
			ID: req.ID,
		}
	}

	if err != nil {
		response = JSONRPCResponse{
			JSONRPC: "2.0",
			Error: &RPCError{
				Code:    -32000,
				Message: err.Error(),
			},
			ID: req.ID,
		}
	}

	// Ensure ID is set
	response.ID = req.ID
	if response.JSONRPC == "" {
		response.JSONRPC = "2.0"
	}

	// Write response
	if useSSE {
		if err := writeSSEResponse(w, response); err != nil {
			log.Printf("Error writing SSE response: %v", err)
		}
	} else {
		if err := writeJSONResponse(w, response); err != nil {
			log.Printf("Error writing JSON response: %v", err)
		}
	}
}

// handleInitialize handles the initialize method
func (s *Server) handleInitialize(req JSONRPCRequest) (JSONRPCResponse, error) {
	result := InitializeResult{
		ProtocolVersion: "2024-11-05",
		Capabilities: map[string]interface{}{
			"tools": true,
		},
		ServerInfo: ServerInfo{
			Name:    "mcp-go",
			Version: "0.1.0",
		},
	}

	return JSONRPCResponse{
		JSONRPC: "2.0",
		Result:  result,
		ID:      req.ID,
	}, nil
}

// handleToolsList handles the tools/list method
func (s *Server) handleToolsList(ctx context.Context, req JSONRPCRequest) (JSONRPCResponse, error) {
	var allTools []interface{}

	// Add local echo tool
	echoTool := tools.GetEchoTool()
	allTools = append(allTools, echoTool)

	// Add local Google PSE tool
	googlePSETool := tools.GetGooglePSETool()
	allTools = append(allTools, googlePSETool)

	// Add tools from gateway (remote MCP servers)
	if s.gateway != nil {
		remoteTools, err := s.gateway.ListAllTools(ctx)
		if err != nil {
			log.Printf("Warning: Failed to list remote tools: %v", err)
		} else {
			// Convert transport.Tool to interface{} for JSON encoding
			for _, tool := range remoteTools {
				allTools = append(allTools, tool)
			}
		}
	}

	result := ToolsListResult{
		Tools: allTools,
	}

	return JSONRPCResponse{
		JSONRPC: "2.0",
		Result:  result,
		ID:      req.ID,
	}, nil
}

// handleToolsCall handles the tools/call method
func (s *Server) handleToolsCall(ctx context.Context, req JSONRPCRequest) (JSONRPCResponse, error) {
	// Extract name and arguments from params
	params := req.Params
	if params == nil {
		return JSONRPCResponse{}, fmt.Errorf("missing params")
	}

	name, ok := params["name"].(string)
	if !ok {
		return JSONRPCResponse{}, fmt.Errorf("missing or invalid 'name' in params")
	}

	arguments, ok := params["arguments"].(map[string]interface{})
	if !ok {
		arguments = make(map[string]interface{})
	}

	// Handle local echo tool
	if name == "echo" {
		message, err := tools.CallEcho(arguments)
		if err != nil {
			return JSONRPCResponse{}, err
		}

		result := ToolCallResult{
			Content: []ContentItem{
				{
					Type: "text",
					Text: message,
				},
			},
		}

		return JSONRPCResponse{
			JSONRPC: "2.0",
			Result:  result,
			ID:      req.ID,
		}, nil
	}

	// Handle local Google PSE tool
	if name == "google_pse_search" {
		result, err := tools.CallGooglePSE(arguments)
		if err != nil {
			return JSONRPCResponse{}, err
		}

		toolResult := ToolCallResult{
			Content: []ContentItem{
				{
					Type: "text",
					Text: result,
				},
			},
		}

		return JSONRPCResponse{
			JSONRPC: "2.0",
			Result:  toolResult,
			ID:      req.ID,
		}, nil
	}

	// Try to handle via gateway (remote MCP servers)
	if s.gateway != nil {
		remoteResp, err := s.gateway.CallTool(ctx, name, arguments)
		if err == nil {
			// Convert transport.ToolResponse to ToolCallResult
			result := ToolCallResult{
				Content: make([]ContentItem, len(remoteResp.Content)),
			}
			for i, item := range remoteResp.Content {
				result.Content[i] = ContentItem{
					Type: item.Type,
					Text: item.Text,
				}
			}

			return JSONRPCResponse{
				JSONRPC: "2.0",
				Result:  result,
				ID:      req.ID,
			}, nil
		}
		// If error is not "not found", return error
		if !isNotFoundError(err) {
			return JSONRPCResponse{}, err
		}
	}

	// Unknown tool
	return JSONRPCResponse{}, fmt.Errorf("tool '%s' not found", name)
}

// isNotFoundError checks if error is a "not found" error
func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	return contains(errMsg, "not found")
}

// contains checks if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Start starts the HTTP server on port 3333
func Start() {
	StartWithGateway(nil)
}

// StartWithGateway starts the HTTP server with a gateway
func StartWithGateway(gw *gateway.Gateway) {
	StartWithGatewayAndPort(gw, ":3333")
}

// StartWithGatewayAndPort starts the HTTP server with a gateway and custom port
func StartWithGatewayAndPort(gw *gateway.Gateway, port string) {
	srv := NewServer(gw)

	// Single MCP endpoint
	http.HandleFunc("/mcp", srv.handleMCP)

	// Also support root path for compatibility
	http.HandleFunc("/", srv.handleMCP)

	// Ensure port starts with ":"
	if port[0] != ':' {
		port = ":" + port
	}

	log.Printf("MCP Server starting on port %s\n", port)
	log.Println("Endpoints available:")
	log.Println("  POST /mcp (JSON-RPC 2.0 over SSE)")
	log.Println("  POST / (JSON-RPC 2.0 over SSE)")
	if gw != nil {
		log.Println("Gateway enabled: Remote MCP servers will be accessible")
	}

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed to start: %v\n", err)
	}
}
