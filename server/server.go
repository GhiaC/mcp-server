package server

import (
	"encoding/json"
	"fmt"
	"log"
	"mcp-go/gateway"
	"mcp-go/tools"
	"net/http"
)

// InitializeResponse represents the response for /initialize endpoint
type InitializeResponse struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    map[string]interface{} `json:"capabilities"`
	ServerInfo      ServerInfo             `json:"serverInfo"`
}

// ServerInfo contains server information
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ToolsListResponse represents the response for /tools/list endpoint
type ToolsListResponse struct {
	Tools []interface{} `json:"tools"` // Can be EchoTool or transport.Tool
}

// Server holds the server state including gateway
type Server struct {
	gateway *gateway.Gateway
}

// NewServer creates a new server instance
func NewServer(gw *gateway.Gateway) *Server {
	return &Server{
		gateway: gw,
	}
}

// ToolCallRequest represents the request for /tools/call endpoint
type ToolCallRequest struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ToolCallResponse represents the response for /tools/call endpoint
type ToolCallResponse struct {
	Content []ContentItem `json:"content"`
}

// ContentItem represents a content item in the tool call response
type ContentItem struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Start starts the HTTP server on port 3333
func Start() {
	StartWithGateway(nil)
}

// StartWithGateway starts the HTTP server with a gateway
func StartWithGateway(gw *gateway.Gateway) {
	srv := NewServer(gw)

	http.HandleFunc("/initialize", handleInitialize)
	http.HandleFunc("/tools/list", srv.handleToolsList)
	http.HandleFunc("/tools/call", srv.handleToolsCall)

	port := ":3333"
	log.Printf("MCP Server starting on port %s\n", port)
	log.Println("Endpoints available:")
	log.Println("  GET  /initialize")
	log.Println("  GET  /tools/list")
	log.Println("  POST /tools/call")
	if gw != nil {
		log.Println("Gateway enabled: Remote MCP servers will be accessible")
	}

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed to start: %v\n", err)
	}
}

// handleInitialize handles GET /initialize
func handleInitialize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := InitializeResponse{
		ProtocolVersion: "2024-11-05",
		Capabilities: map[string]interface{}{
			"tools": true,
		},
		ServerInfo: ServerInfo{
			Name:    "mcp-go",
			Version: "0.1.0",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
		return
	}
}

// handleToolsList handles GET /tools/list
func (s *Server) handleToolsList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var allTools []interface{}

	// Add local echo tool
	echoTool := tools.GetEchoTool()
	allTools = append(allTools, echoTool)

	// Add local Google PSE tool
	googlePSETool := tools.GetGooglePSETool()
	allTools = append(allTools, googlePSETool)

	// Add tools from gateway (remote MCP servers)
	if s.gateway != nil {
		ctx := r.Context()
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

	response := ToolsListResponse{
		Tools: allTools,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
		return
	}
}

// handleToolsCall handles POST /tools/call
func (s *Server) handleToolsCall(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ToolCallRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Error decoding request: %v", err), http.StatusBadRequest)
		return
	}

	// Handle local echo tool
	if req.Name == "echo" {
		message, err := tools.CallEcho(req.Arguments)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error calling echo tool: %v", err), http.StatusBadRequest)
			return
		}

		response := ToolCallResponse{
			Content: []ContentItem{
				{
					Type: "text",
					Text: message,
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
			return
		}
		return
	}

	// Handle local Google PSE tool
	if req.Name == "google_pse_search" {
		result, err := tools.CallGooglePSE(req.Arguments)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error calling Google PSE tool: %v", err), http.StatusBadRequest)
			return
		}

		response := ToolCallResponse{
			Content: []ContentItem{
				{
					Type: "text",
					Text: result,
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
			return
		}
		return
	}

	// Try to handle via gateway (remote MCP servers)
	if s.gateway != nil {
		ctx := r.Context()
		remoteResp, err := s.gateway.CallTool(ctx, req.Name, req.Arguments)
		if err == nil {
			// Convert transport.ToolResponse to ToolCallResponse
			response := ToolCallResponse{
				Content: make([]ContentItem, len(remoteResp.Content)),
			}
			for i, item := range remoteResp.Content {
				response.Content[i] = ContentItem{
					Type: item.Type,
					Text: item.Text,
				}
			}

			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(response); err != nil {
				http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
				return
			}
			return
		}
		// If error is not "not found", return error
		if !isNotFoundError(err) {
			http.Error(w, fmt.Sprintf("Error calling remote tool: %v", err), http.StatusInternalServerError)
			return
		}
	}

	// Unknown tool
	http.Error(w, "Tool not found", http.StatusNotFound)
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
