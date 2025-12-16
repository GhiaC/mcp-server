package main

import (
	"encoding/json"
	"fmt"
	"log"
	"mcp-go/server"
	"mcp-go/tools"
	"net/http"
)

func main() {
	// Create a simple server for filesystem operations
	srv := NewFileSystemServer()

	http.HandleFunc("/initialize", handleInitialize)
	http.HandleFunc("/tools/list", srv.handleToolsList)
	http.HandleFunc("/tools/call", srv.handleToolsCall)

	port := ":3335"
	log.Printf("FileSystem MCP Server starting on port %s\n", port)
	log.Println("Endpoints available:")
	log.Println("  GET  /initialize")
	log.Println("  GET  /tools/list")
	log.Println("  POST /tools/call")

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed to start: %v\n", err)
	}
}

// FileSystemServer handles filesystem MCP operations
type FileSystemServer struct{}

func NewFileSystemServer() *FileSystemServer {
	return &FileSystemServer{}
}

// handleInitialize handles GET /initialize
func handleInitialize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := server.InitializeResponse{
		ProtocolVersion: "2024-11-05",
		Capabilities: map[string]interface{}{
			"tools": true,
		},
		ServerInfo: server.ServerInfo{
			Name:    "filesystem-mcp",
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
func (s *FileSystemServer) handleToolsList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var allTools []interface{}

	// Add filesystem tools with "filesystem:" prefix
	readFileTool := tools.GetReadFileTool()
	readFileTool.Name = "filesystem:read_file"
	allTools = append(allTools, readFileTool)

	writeFileTool := tools.GetWriteFileTool()
	writeFileTool.Name = "filesystem:write_file"
	allTools = append(allTools, writeFileTool)

	listDirTool := tools.GetListDirectoryTool()
	listDirTool.Name = "filesystem:list_directory"
	allTools = append(allTools, listDirTool)

	createDirTool := tools.GetCreateDirectoryTool()
	createDirTool.Name = "filesystem:create_directory"
	allTools = append(allTools, createDirTool)

	deleteFileTool := tools.GetDeleteFileTool()
	deleteFileTool.Name = "filesystem:delete_file"
	allTools = append(allTools, deleteFileTool)

	response := server.ToolsListResponse{
		Tools: allTools,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
		return
	}
}

// handleToolsCall handles POST /tools/call
func (s *FileSystemServer) handleToolsCall(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req server.ToolCallRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Error decoding request: %v", err), http.StatusBadRequest)
		return
	}

	var result string
	var err error

	// Handle filesystem tools
	switch req.Name {
	case "filesystem:read_file":
		result, err = tools.CallReadFile(req.Arguments)
	case "filesystem:write_file":
		result, err = tools.CallWriteFile(req.Arguments)
	case "filesystem:list_directory":
		result, err = tools.CallListDirectory(req.Arguments)
	case "filesystem:create_directory":
		result, err = tools.CallCreateDirectory(req.Arguments)
	case "filesystem:delete_file":
		result, err = tools.CallDeleteFile(req.Arguments)
	default:
		http.Error(w, "Tool not found", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, fmt.Sprintf("Error calling tool: %v", err), http.StatusBadRequest)
		return
	}

	response := server.ToolCallResponse{
		Content: []server.ContentItem{
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
}
