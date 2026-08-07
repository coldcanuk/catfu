# catfu MCP tools

Start the server:

```bash
catfu mcp
# or with explicit DB / keys:
CATFU_DB=~/.local/share/catfu/catfu.db BRAVE_API_KEY=… catfu mcp
```

Transport: **stdio** (JSON-RPC). Logs go to **stderr**.

## Tools

| Tool | Description | Key inputs |
|------|-------------|------------|
| `catalogue_channel` | Ingest channel metadata via yt-dlp | `channel_url`, optional `full`, `limit` |
| `search_catalogue` | Local FTS + filters | `query`, `channel_id`, `after`, `before`, `limit`, `offset` |
| `web_search` | Brave Search plan (web/news/video) | `query`, `kind`, `limit`, `offset`, `country`, `freshness`, … |
| `list_catalogues` | List stored channels | (none) |
| `doctor` | Health / dependency report | (none) |

All tools return JSON text content suitable for agents.

## Example client config (Claude Desktop style)

```json
{
  "mcpServers": {
    "catfu": {
      "command": "catfu",
      "args": ["mcp"],
      "env": {
        "CATFU_DB": "/path/to/catfu.db",
        "BRAVE_API_KEY": "optional"
      }
    }
  }
}
```
