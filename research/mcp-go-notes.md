# MCP Go SDK research notes (catfu)

**Date:** 2026-08-07  
**Chosen library:** official `github.com/modelcontextprotocol/go-sdk` (module path; import `.../mcp`)  
**Version pin target:** latest stable at build time (documented v1.x, e.g. v1.7.0+)  
**Protocol:** 2025–2026 MCP revisions supported by official SDK (including 2026-07-28 docs family)

## Server bootstrap (stdio)

```go
server := mcp.NewServer(&mcp.Implementation{
    Name:    "catfu",
    Version: version.Version,
}, nil)

mcp.AddTool(server, &mcp.Tool{
    Name:        "search_catalogue",
    Description: "Full-text search local YouTube catalogue",
}, searchHandler)

if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
    log.Fatal(err)
}
```

## Typed tools

```go
type SearchArgs struct {
    Query  string `json:"query" jsonschema:"search query"`
    Limit  int    `json:"limit,omitempty" jsonschema:"max results"`
    Offset int    `json:"offset,omitempty"`
}

func searchHandler(ctx context.Context, req *mcp.CallToolRequest, args SearchArgs) (*mcp.CallToolResult, any, error) {
    // return text content + optional structured payload
}
```

`AddTool` infers JSON Schema from struct tags; validation is automatic.

## Transport

- **stdio only** for v1 foundation (`mcp.StdioTransport{}`).
- Logging must go to **stderr** so stdout stays clean for JSON-RPC.

## Tools to expose (Milestone 9.1)

| Tool | Purpose |
|------|---------|
| `catalogue_channel` | ingest channel metadata via yt-dlp |
| `search_catalogue` | local FTS + date/channel filters |
| `web_search` | Brave backend |
| `list_catalogues` | list stored channels |
| `status` / `doctor` | health / dependency checks |

## Error handling for agents

- Return errors from handlers (SDK marks `IsError`).
- Prefer structured JSON in text content for agent parseability.
- Never put secrets in tool responses.

## BYOK

MCP host injects env (`BRAVE_API_KEY`, `CATFU_DB`, etc.) into the process; catfu reads via config layer.
