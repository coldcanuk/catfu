# Brave Search API research notes (catfu)

**Updated:** 2026-08-07  
**Docs:** https://api-dashboard.search.brave.com/documentation

## Which key / subscription?

| Product | Use for catfu? | Notes |
|---------|----------------|-------|
| **Search** | **Yes — required** | One `X-Subscription-Token` covers Web, News, Video, Image, LLM Context |
| **Answers** | No (optional later) | Separate plan; chat/summarizer billing |

Pricing (2026 public): Search ≈ **$5 / 1 000 requests**, **$5 free monthly credits**, ~50 rps.

Multiple dashboard keys are multiple tokens for the same product — catfu needs **one Search key**:

| Source | Name | Notes |
|--------|------|-------|
| **Keychain / secrets file** | `catfu auth set` | Preferred for humans |
| Env | `BRAVE_API_KEY` or `CATFU_BRAVE_API_KEY` | CI / agents |
| Flag | `--brave-api-key` | One-off override |
| Config | `brave_api_key` | Discouraged if file is shared/backed up naively |

## Auth

```
X-Subscription-Token: <SEARCH_PLAN_TOKEN>
Accept: application/json
```

Do **not** set `Accept-Encoding: gzip` in application code; let `net/http` Transport negotiate and decompress.

## Endpoints used by catfu

| Kind | URL | count max |
|------|-----|-----------|
| web | `GET https://api.search.brave.com/res/v1/web/search` | 20 |
| news | `GET https://api.search.brave.com/res/v1/news/search` | 50 |
| video | `GET https://api.search.brave.com/res/v1/videos/search` | 50 |

Common params: `q`, `count`, `offset` (0–9), `freshness`, `country`, `search_lang`, `safesearch`.

## Rate limits

Headers: `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`, `X-RateLimit-Policy`.  
On 429, surface reset seconds to the user.

## Out of scope for foundation

Answers / Summarizer streaming API, Image Search, Place Search, LLM Context endpoint (same key can be added later).
