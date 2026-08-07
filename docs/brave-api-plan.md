# Plan: Brave Search API correctness for catfu

## Methodology: Risk-Driven Phased Planning (Spiral-adapted)

**Why:** Brave API surface is multi-product (Search vs Answers), multi-endpoint
(web/news/video/…), rate-limited, and subscription-gated. Highest risk is
external contract drift and silent client bugs (auth, gzip, limits). Research
cycles first, then incremental delivery with tests, reduces rework.

**Evidence for methodology choice:**
- Multiple endpoints with different count limits (web max 20, video/news max 50)
- Rate-limit headers (`X-RateLimit-*`) and 429 semantics require careful handling
- Dashboard issues product-scoped subscription tokens (not one universal free key)
- Prior client set `Accept-Encoding: gzip` (Go does not auto-decompress when set → JSON decode failures)

## Scope

### In scope
- Audit + fix Brave HTTP client (auth, gzip, params, errors, rate-limit headers)
- Web, News, and Video search endpoints under one Search-plan API key
- CLI flags for kind/country/lang/safesearch/freshness
- MCP `web_search` richer inputs; optional news/video via kind
- Config/docs: which subscription/key, env vars, doctor reporting
- Unit tests with httptest
- Research notes + ADR

### Out of scope (this plan)
- Answers / Summarizer streaming chat API (separate product pricing)
- Image Search, Place Search, LLM Context (can share client later)
- Paid live integration tests requiring a real key in CI

## Recommended subscription & keys

| Product | Covers | Best for catfu? |
|---------|--------|-----------------|
| **Search** | Web, News, Video, Image, LLM Context | **Yes — required** |
| **Answers** | Grounded answer generation + token billing | No (optional later) |

- Use **one** `X-Subscription-Token` from a **Search** plan subscription.
- Pricing (2026): ~$5 / 1k requests, **$5 free monthly credits**, ~50 rps capacity.
- Multiple dashboard keys are just multiple tokens for the same product; catfu needs **one** Search key via `BRAVE_API_KEY` / `--brave-api-key` / config.

## Phases

### Phase 1 — Research & Risk Resolution
#### Milestone 1.1 — External API research
#### Milestone 1.2 — Codebase audit & plan update
### Phase 2 — Client core fix
#### Milestone 2.1 — HTTP client correctness (gzip, headers, errors)
#### Milestone 2.2 — Multi-endpoint (web/news/video) + query options
### Phase 3 — CLI / MCP / Config integration
#### Milestone 3.1 — CLI + config flags
#### Milestone 3.2 — MCP tools + doctor
### Phase 4 — Docs, tests polish, release
#### Milestone 4.1 — Tests + docs
#### Milestone 4.2 — Audit, merge main, remove worktree

Worktree: `wt/brave-api-v2`

## Plan update (end of Phase 1 / Milestone 1.2)

### Concrete decisions
1. **Subscription:** Brave **Search** plan (not Answers). One product key.
2. **Env:** keep `BRAVE_API_KEY` / `CATFU_BRAVE_API_KEY` / `brave_api_key` / `--brave-api-key`.
3. **Endpoints:** web (default), news, video via `--kind` / MCP `kind`.
4. **Fix:** remove manual `Accept-Encoding: gzip` (let Transport handle).
5. **Params:** country, search_lang, safesearch, freshness shortcuts + date range.
6. **Tests:** httptest for web/news/video decode paths and 401/429.
7. **Docs:** README + research notes + doctor hint for Search plan.

### Revised estimates
- Client rewrite + tests: small–medium
- CLI/MCP wiring: small
- Docs: small
