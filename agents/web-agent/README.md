# Web Agent for AgentForgeEngine

A token-optimized web fetching agent that extracts clean, LLM-friendly content from web pages.

## Features

- **Token-Optimized Output**: Extracts meaningful content, not raw HTML
- **Smart Content Extraction**: Focuses on main content, removes boilerplate
- **Configurable Token Limits**: Default 8k tokens with per-request override
- **Domain Filtering**: Allowlist/blocklist support for security
- **Content Type Validation**: Only processes safe content types
- **Rich Metadata**: Title, description, headings, links, word/token counts
- **Graceful Error Handling**: Partial extraction with warnings
- **Hot Reload Ready**: Supports seamless replacement with `agentforge reload`

## Operations

### `fetch`
Fetch and extract content from a URL.

**Input:**
```json
{
  "type": "fetch",
  "payload": {
    "url": "https://example.com/article",
    "max_tokens": 4000
  }
}
```

**Output:**
```json
{
  "success": true,
  "data": {
    "url": "https://example.com/article",
    "title": "Article Title",
    "description": "Brief description",
    "main_content": "Clean article text...",
    "headings": [
      {"level": 1, "text": "Introduction"},
      {"level": 2, "text": "Main Content"}
    ],
    "links": [
      {"text": "Related Article", "url": "/related", "type": "internal"}
    ],
    "metadata": {
      "content_length": "15000",
      "content_density": "high"
    },
    "token_count": 3847,
    "word_count": 1200,
    "truncated": false
  }
}
```

### `validate`
Check if a URL is accessible and allowed without downloading content.

**Input:**
```json
{
  "type": "validate",
  "payload": {
    "url": "https://example.com"
  }
}
```

### `extract`
Alias for `fetch` operation (can be enhanced with custom extraction logic).

## Configuration

Add to your `agentforge.yaml`:

```yaml
agents:
  local:
    - name: "web-agent"
      path: "./agents/web-agent"
      config:
        default_max_tokens: 8000
        max_allowed_tokens: 15000
        min_allowed_tokens: 500
        timeout: 15
        user_agent: "AgentForgeEngine-WebAgent/1.0"
        allowed_domains: ["*"]
        blocked_domains: ["ads.*", "trackers.*"]
        content_types: ["text/html", "application/json", "text/plain"]
        include_links: true
        include_metadata: true
```

## Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `default_max_tokens` | int | 8000 | Default token limit for content |
| `max_allowed_tokens` | int | 15000 | Maximum allowed tokens (safety limit) |
| `min_allowed_tokens` | int | 500 | Minimum allowed tokens |
| `timeout` | int | 15 | Request timeout in seconds |
| `user_agent` | string | "AgentForgeEngine-WebAgent/1.0" | HTTP User-Agent header |
| `allowed_domains` | array | ["*"] | Allowed domains (wildcards supported) |
| `blocked_domains` | array | [] | Blocked domains (wildcards supported) |
| `content_types` | array | ["text/html", "text/plain", "application/json"] | Allowed content types |
| `include_links` | bool | true | Extract links from pages |
| `include_metadata` | bool | true | Include extraction metadata |

## Content Extraction Strategy

The agent prioritizes content for LLM consumption:

1. **High Priority**: Title, meta description, main content
2. **Medium Priority**: Headings, important links
3. **Low Priority**: Metadata, statistics

### Token Budget Breakdown (8000 tokens)
- Title: 100 tokens
- Description: 300 tokens  
- Headings: 400 tokens
- Links: 200 tokens
- Main Content: 7000 tokens

## Hot Reload Usage

Replace the agent without downtime:

```bash
# Create improved version
mkdir -p custom-agents/web-agent-v2
# Build your custom agent...

# Update configuration
# path: "./custom-agents/web-agent-v2"

# Reload
agentforge reload --agent web-agent
```

## Security Features

- Content size limits (10MB max download)
- Domain filtering (allowlist/blocklist)
- Content type validation
- Automatic boilerplate removal
- Smart truncation for token limits

## Error Handling

- Network timeouts handled gracefully
- Invalid URLs return clear error messages
- Partial extraction attempts with warnings
- HTTP status codes properly handled
- Content type violations rejected

## Development

Built with Go standard library + `golang.org/x/net/html` for lightweight dependency footprint.

### Testing
```bash
cd agents/web-agent
go build -buildmode=plugin -o ../plugins/web-agent.so .
```