package tools

import (
	"fmt"
	"os"
	"path/filepath"
)

// FileSystemTool represents a filesystem tool definition
type FileSystemTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// GetReadFileTool returns the read_file tool definition
func GetReadFileTool() FileSystemTool {
	return FileSystemTool{
		Name:        "read_file",
		Description: "Read the contents of a file",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "The path to the file to read",
				},
			},
			"required": []string{"path"},
		},
	}
}

// GetWriteFileTool returns the write_file tool definition
func GetWriteFileTool() FileSystemTool {
	return FileSystemTool{
		Name:        "write_file",
		Description: "Write content to a file",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "The path to the file to write",
				},
				"content": map[string]interface{}{
					"type":        "string",
					"description": "The content to write to the file",
				},
			},
			"required": []string{"path", "content"},
		},
	}
}

// GetListDirectoryTool returns the list_directory tool definition
func GetListDirectoryTool() FileSystemTool {
	return FileSystemTool{
		Name:        "list_directory",
		Description: "List files and directories in a directory",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "The path to the directory to list",
				},
			},
			"required": []string{"path"},
		},
	}
}

// GetCreateDirectoryTool returns the create_directory tool definition
func GetCreateDirectoryTool() FileSystemTool {
	return FileSystemTool{
		Name:        "create_directory",
		Description: "Create a new directory",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "The path to the directory to create",
				},
			},
			"required": []string{"path"},
		},
	}
}

// GetDeleteFileTool returns the delete_file tool definition
func GetDeleteFileTool() FileSystemTool {
	return FileSystemTool{
		Name:        "delete_file",
		Description: "Delete a file or directory",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "The path to the file or directory to delete",
				},
			},
			"required": []string{"path"},
		},
	}
}

// CallReadFile reads a file and returns its contents
func CallReadFile(arguments map[string]interface{}) (string, error) {
	path, ok := arguments["path"].(string)
	if !ok {
		return "", fmt.Errorf("path argument is required and must be a string")
	}

	// Resolve absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path: %v", err)
	}

	content, err := os.ReadFile(absPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %v", err)
	}

	return string(content), nil
}

// CallWriteFile writes content to a file
func CallWriteFile(arguments map[string]interface{}) (string, error) {
	path, ok := arguments["path"].(string)
	if !ok {
		return "", fmt.Errorf("path argument is required and must be a string")
	}

	content, ok := arguments["content"].(string)
	if !ok {
		return "", fmt.Errorf("content argument is required and must be a string")
	}

	// Resolve absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path: %v", err)
	}

	// Create parent directories if they don't exist
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create parent directories: %v", err)
	}

	if err := os.WriteFile(absPath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %v", err)
	}

	return fmt.Sprintf("Successfully wrote %d bytes to %s", len(content), absPath), nil
}

// CallListDirectory lists files and directories in a directory
func CallListDirectory(arguments map[string]interface{}) (string, error) {
	path, ok := arguments["path"].(string)
	if !ok {
		return "", fmt.Errorf("path argument is required and must be a string")
	}

	// Resolve absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path: %v", err)
	}

	entries, err := os.ReadDir(absPath)
	if err != nil {
		return "", fmt.Errorf("failed to read directory: %v", err)
	}

	result := fmt.Sprintf("Contents of %s:\n", absPath)
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		entryType := "file"
		if entry.IsDir() {
			entryType = "directory"
		}
		result += fmt.Sprintf("  %s [%s] %d bytes\n", entry.Name(), entryType, info.Size())
	}

	return result, nil
}

// CallCreateDirectory creates a new directory
func CallCreateDirectory(arguments map[string]interface{}) (string, error) {
	path, ok := arguments["path"].(string)
	if !ok {
		return "", fmt.Errorf("path argument is required and must be a string")
	}

	// Resolve absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path: %v", err)
	}

	if err := os.MkdirAll(absPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %v", err)
	}

	return fmt.Sprintf("Successfully created directory: %s", absPath), nil
}

// CallDeleteFile deletes a file or directory
func CallDeleteFile(arguments map[string]interface{}) (string, error) {
	path, ok := arguments["path"].(string)
	if !ok {
		return "", fmt.Errorf("path argument is required and must be a string")
	}

	// Resolve absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path: %v", err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return "", fmt.Errorf("file or directory does not exist: %v", err)
	}

	if info.IsDir() {
		if err := os.RemoveAll(absPath); err != nil {
			return "", fmt.Errorf("failed to delete directory: %v", err)
		}
		return fmt.Sprintf("Successfully deleted directory: %s", absPath), nil
	}

	if err := os.Remove(absPath); err != nil {
		return "", fmt.Errorf("failed to delete file: %v", err)
	}

	return fmt.Sprintf("Successfully deleted file: %s", absPath), nil
}
