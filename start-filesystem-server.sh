#!/bin/bash

# Start the filesystem MCP server on port 3335
echo "Starting FileSystem MCP Server on port 3335..."
go run ./cmd/filesystem-server/main.go
