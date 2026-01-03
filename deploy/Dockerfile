# Use a minimal image for the pre-built MCP Server binary
FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the pre-built binary and configuration
COPY deploy/mcp-server .
COPY mcp-config.json mcp-config.json

EXPOSE 3333

CMD ["./mcp-server"]
