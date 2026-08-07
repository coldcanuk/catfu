# Brave Search API audit (2026-08-07)

Sources:
- https://api-dashboard.search.brave.com/documentation
- Web Search GET reference
- Video Search service + GET reference
- News Search service
- Rate limiting guide
- Pricing / Search vs Answers plans

## Products vs endpoints

Brave sells **plans/products**. The important split:

1. **Search plan** — data APIs: Web, News, Video, Image, Place, LLM Context  
   - Auth: `X-Subscription-Token: <search_product_key>`  
   - Price: ~$5 / 1k requests; $5 free monthly credits; ~50 rps

2. **Answers plan** — AI answers (chat-style / summarizer)  
   - Separate product/token model and token billing  
   - **Not required for catfu v1 web/news/video search**

Dashboard may let you create multiple keys per product; they are not different
API types. catfu should document **one Search subscription key**.

## Endpoints (base host `https://api.search.brave.com`)

| Service | Path | count max | Notes |
|---------|------|-----------|-------|
| Web | `/res/v1/web/search` | 20 | Response nest: `web.results[]` |
| News | `/res/v1/news/search` | 50 | `results[]` (or news-shaped) |
| Video | `/res/v1/videos/search` | 50 | `type: videos`, `results[]` |

Auth header (all): **`X-Subscription-Token`**  
Accept: `application/json`

## Web query params (high value for catfu)

- `q` (required, max 400 chars / 50 words)
- `count` 1–20, `offset` 0–9
- `freshness`: `pd|pw|pm|py` or `YYYY-MM-DDtoYYYY-MM-DD`
- `country`, `search_lang`, `safesearch` (off|moderate|strict)
- `extra_snippets` (plan-dependent)

## Rate limiting

Headers: `X-RateLimit-Limit`, `X-RateLimit-Policy`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`  
429 → respect reset; exponential backoff. Only successful requests count.

## Issues found in current catfu client

1. **Critical:** `Accept-Encoding: gzip` is set manually → Go transport will NOT
   auto-decompress → body may be gzip binary → JSON unmarshal fails in production.
2. Only web endpoint; no news/video despite Search plan including them.
3. Missing country/lang/safesearch flags (API supports them).
4. 429 error ignores `X-RateLimit-Reset`.
5. No httptest tests.
6. Docs under-specified subscription recommendation.
