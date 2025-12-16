package tools

import "fmt"

// EchoTool represents the echo tool definition
type EchoTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// GetEchoTool returns the echo tool definition
func GetEchoTool() EchoTool {
	return EchoTool{
		Name:        "echo",
		Description: "Echo back the provided message",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"message": map[string]interface{}{
					"type":        "string",
					"description": "The message to echo back",
				},
			},
			"required": []string{"message"},
		},
	}
}

// CallEcho executes the echo tool with the given arguments
func CallEcho(arguments map[string]interface{}) (string, error) {
	message, ok := arguments["message"].(string)
	if !ok {
		return "", fmt.Errorf("message argument is required and must be a string")
	}
	return message, nil
}
