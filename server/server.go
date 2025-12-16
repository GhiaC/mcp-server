package server

import (
	"encoding/json"
	"fmt"
	"log"
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
	Tools []tools.EchoTool `json:"tools"`
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
	http.HandleFunc("/initialize", handleInitialize)
	http.HandleFunc("/tools/list", handleToolsList)
	http.HandleFunc("/tools/call", handleToolsCall)

	port := ":3333"
	log.Printf("MCP Server starting on port %s\n", port)
	log.Println("Endpoints available:")
	log.Println("  GET  /initialize")
	log.Println("  GET  /tools/list")
	log.Println("  POST /tools/call")

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
func handleToolsList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	echoTool := tools.GetEchoTool()
	response := ToolsListResponse{
		Tools: []tools.EchoTool{echoTool},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
		return
	}
}

// handleToolsCall handles POST /tools/call
func handleToolsCall(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ToolCallRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Error decoding request: %v", err), http.StatusBadRequest)
		return
	}

	// Handle echo tool
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

	// Unknown tool
	http.Error(w, "Tool not found", http.StatusNotFound)
}
