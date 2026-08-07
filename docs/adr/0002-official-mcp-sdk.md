# ADR 0002: Official MCP Go SDK

## Status
Accepted

## Context
Agents need a stdio MCP server. Community SDKs vary in protocol support.

## Decision
Use `github.com/modelcontextprotocol/go-sdk` with `mcp.NewServer`, typed `mcp.AddTool`, and `mcp.StdioTransport`.

## Consequences
- Typed tool schemas from Go structs.
- Logging must use stderr.
- Pin a tested module version in go.mod; upgrade deliberately when protocol changes.
