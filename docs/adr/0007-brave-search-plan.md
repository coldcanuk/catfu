# ADR 0007: Brave Search plan single-key client

## Status
Accepted

## Context
Brave offers Search and Answers products with separate subscription tokens.
catfu needs public search for agents/CLI (web, optionally news/video), not grounded answer generation.

## Decision
1. Require a **Search** plan `X-Subscription-Token` via `BRAVE_API_KEY`.
2. Support **web**, **news**, and **video** endpoints with one client.
3. Do not integrate Answers/Summarizer in this iteration.
4. Let Go's HTTP Transport manage gzip (never set Accept-Encoding in app code).

## Consequences
- Users must subscribe to Search (monthly free credits exist).
- Answers-only keys will fail with auth/plan errors — documented.
- Future LLM Context can reuse the same key and client patterns.
