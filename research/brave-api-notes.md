# Brave Search API research notes (catfu)

**Date:** 2026-08-07  
**Primary docs:** https://api-dashboard.search.brave.com/app/documentation/web-search/get-started

## Auth

- Header: `X-Subscription-Token: <API_KEY>`
- Also send `Accept: application/json`
- Optional: `Accept-Encoding: gzip`

## Endpoint (Web Search)

```
GET https://api.search.brave.com/res/v1/web/search
```

### Query parameters (subset used by catfu)

| Param | Notes |
|-------|-------|
| `q` | required query |
| `count` | max 20 per page (default 20) |
| `offset` | 0-based, max 9 (API constraint) |
| `freshness` | `pd` / `pw` / `pm` / `py` or `YYYY-MM-DDtoYYYY-MM-DD` |
| `country` | optional 2-letter code |
| `search_lang` | optional |
| `safesearch` | `off` / `moderate` / `strict` |

## Response shape (relevant)

```json
{
  "query": { "original": "...", "more_results_available": true },
  "web": {
    "results": [
      {
        "title": "...",
        "url": "...",
        "description": "...",
        "age": "...",
        "extra_snippets": ["..."]
      }
    ]
  }
}
```

Map to catfu `Result`: title, url, description/snippet, source=`brave`.

## Rate limits / quotas

- Enforced with short sliding windows; responses may include `X-RateLimit-*` headers.
- Free/trial tiers are modest; treat 429 as retryable with backoff.
- Default capacity often cited ~50 rps for paid plans — catfu is CLI/MCP and stays far below this.

## Client approach

- **Thin custom HTTP client** in `internal/backends/brave` (stdlib `net/http`).
- No incomplete third-party wrapper dependency.
- Timeouts via `context.Context` + `http.Client{Timeout}`.
- Never log the API key (redact in `catfu config` / doctor).

## Credentials (BYOK)

| Source | Name |
|--------|------|
| Env | `BRAVE_API_KEY` or `CATFU_BRAVE_API_KEY` |
| Flag | `--brave-api-key` |
| Config file | `brave_api_key` |

## Errors

| HTTP | Handling |
|------|----------|
| 401 / 403 | invalid/missing key — non-retryable |
| 429 | rate limited — retry with backoff, surface clear message |
| 5xx | retry once/twice then fail |
| network | context-aware error |
