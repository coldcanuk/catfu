# Codebase audit: internal/backends/brave

## Current shape
- Single `Client.Search` → GET web/search only
- Maps `web.results` → backends.Result
- Config: `BraveAPIKey` / `BRAVE_API_KEY` / `--brave-api-key`

## Verdict
Auth URL and token header are **correct for Web Search**.
Implementation is **incomplete and has a gzip bug**. Not production-safe.

## Required changes (plan Phase 2–4)
- Fix compression handling
- Kind: web | news | video
- Richer SearchQuery options or Brave-specific Options
- RateLimit metadata in errors
- Tests with httptest
- CLI/MCP/docs alignment
