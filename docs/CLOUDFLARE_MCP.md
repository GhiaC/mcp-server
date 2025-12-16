# Cloudflare MCP Servers Integration

This document describes how to integrate Cloudflare's MCP servers with your MCP Go server.

## Available Cloudflare MCP Servers

Cloudflare provides multiple MCP servers for different use cases. All servers use the `streamable-http` transport via `/mcp` endpoint.

### Server List

| Server Name | URL | Description | Auth Required |
|------------|-----|-------------|---------------|
| **Documentation** | `https://docs.mcp.cloudflare.com/mcp` | Get up to date reference information on Cloudflare | No |
| **Workers Bindings** | `https://bindings.mcp.cloudflare.com/mcp` | Build Workers applications with storage, AI, and compute primitives | Yes |
| **Workers Builds** | `https://builds.mcp.cloudflare.com/mcp` | Get insights and manage your Cloudflare Workers Builds | Yes |
| **Observability** | `https://observability.mcp.cloudflare.com/mcp` | Debug and get insight into your application's logs and analytics | Yes |
| **Radar** | `https://radar.mcp.cloudflare.com/mcp` | Get global Internet traffic insights, trends, URL scans | No |
| **Container** | `https://containers.mcp.cloudflare.com/mcp` | Spin up a sandbox development environment | Yes |
| **Browser Rendering** | `https://browser.mcp.cloudflare.com/mcp` | Fetch web pages, convert to markdown and take screenshots | Yes |
| **Logpush** | `https://logs.mcp.cloudflare.com/mcp` | Get quick summaries for Logpush job health | Yes |
| **AI Gateway** | `https://ai-gateway.mcp.cloudflare.com/mcp` | Search your logs, get details about prompts and responses | Yes |
| **AutoRAG** | `https://autorag.mcp.cloudflare.com/mcp` | List and search documents on your AutoRAGs | Yes |
| **Audit Logs** | `https://auditlogs.mcp.cloudflare.com/mcp` | Query audit logs and generate reports | Yes |
| **DNS Analytics** | `https://dns-analytics.mcp.cloudflare.com/mcp` | Optimize DNS performance and debug issues | Yes |
| **Digital Experience Monitoring** | `https://dex.mcp.cloudflare.com/mcp` | Get insight on critical applications | Yes |
| **Cloudflare One CASB** | `https://casb.mcp.cloudflare.com/mcp` | Identify security misconfigurations for SaaS applications | Yes |
| **GraphQL** | `https://graphql.mcp.cloudflare.com/mcp` | Get analytics data using Cloudflare's GraphQL API | Yes |

## Transport Protocol

All Cloudflare MCP servers use the **`streamable-http`** transport protocol via the `/mcp` endpoint. This is compatible with the standard HTTP transport implementation in this MCP Go server.

## Repository Reference

For the latest information, updates, and source code, visit:
- **GitHub Repository**: [cloudflare/mcp-server-cloudflare](https://github.com/cloudflare/mcp-server-cloudflare)
- **Stars**: 3.2k+ ⭐
- **License**: Apache-2.0

## Configuration

### 1. Get Cloudflare API Token

For servers that require authentication, you need a Cloudflare API token:

1. Go to [Cloudflare Dashboard](https://dash.cloudflare.com/profile/api-tokens)
2. Create a new API token with appropriate scopes
3. Copy the token

### 2. Configure in mcp-config.json

Add Cloudflare servers to your `mcp-config.json`:

```json
{
  "servers": [
    {
      "name": "cloudflare-observability",
      "url": "https://observability.mcp.cloudflare.com/mcp",
      "transport": "http",
      "enabled": true,
      "prefix": "cloudflare:",
      "auth": {
        "Authorization": "Bearer YOUR_CLOUDFLARE_API_TOKEN"
      }
    },
    {
      "name": "cloudflare-docs",
      "url": "https://docs.mcp.cloudflare.com/mcp",
      "transport": "http",
      "enabled": true,
      "prefix": "cloudflare-docs:",
      "auth": {}
    }
  ]
}
```

### 3. Environment Variables (Alternative)

You can also configure via environment variables:

```bash
export CLOUDFLARE_API_TOKEN="your-api-token"
```

## Usage Examples

### Using Observability Server

```bash
# List available tools
curl http://localhost:3333/tools/list | jq '.tools[] | select(.name | startswith("cloudflare:"))'

# Call a tool (example)
curl -X POST http://localhost:3333/tools/call \
  -H "Content-Type: application/json" \
  -d '{
    "name": "cloudflare:get_logs",
    "arguments": {
      "account_id": "your-account-id",
      "start_time": "2024-01-01T00:00:00Z"
    }
  }'
```

### Using Documentation Server

```bash
# Search Cloudflare documentation
curl -X POST http://localhost:3333/tools/call \
  -H "Content-Type: application/json" \
  -d '{
    "name": "cloudflare-docs:search",
    "arguments": {
      "query": "Workers API"
    }
  }'
```

### Using Radar Server

```bash
# Get Internet traffic insights
curl -X POST http://localhost:3333/tools/call \
  -H "Content-Type: application/json" \
  -d '{
    "name": "cloudflare-radar:get_traffic_insights",
    "arguments": {
      "location": "US"
    }
  }'
```

## API Token Scopes

Different Cloudflare MCP servers require different API token scopes. When creating your API token in the Cloudflare dashboard, ensure you grant the necessary permissions:

- **Workers**: `Workers:Read`, `Workers:Write`
- **Analytics**: `Analytics:Read`
- **Logs**: `Logs:Read`
- **DNS**: `DNS:Read`, `DNS:Write`
- **Account**: `Account:Read`

## Troubleshooting

### Authentication Errors

If you get authentication errors:
1. Verify your API token is correct
2. Check that the token has the required scopes
3. Ensure the token hasn't expired

### Connection Errors

If you can't connect to Cloudflare servers:
1. Check your internet connection
2. Verify the server URL is correct
3. Check Cloudflare's status page for outages

### Rate Limiting

Cloudflare MCP servers may have rate limits. If you hit limits:
- Reduce the frequency of requests
- Use caching where possible
- Contact Cloudflare support for higher limits

## References

- [Cloudflare MCP Server Repository](https://github.com/cloudflare/mcp-server-cloudflare)
- [Cloudflare API Documentation](https://developers.cloudflare.com/api/)
- [MCP Protocol Specification](https://modelcontextprotocol.io/)

