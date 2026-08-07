# ADR 0009: Brave as catalogue force multiplier

## Status
Accepted

## Context
catfu's core product is a **local YouTube channel metadata catalogue**. Brave Search
was integrated as a parallel BYOK web backend. Users correctly asked how Brave
should amplify cataloguing rather than sit beside it.

## Decision
1. **`discover`** — Brave video + `site:youtube.com` web → extract channel/video
   identities → correlate with SQLite (already catalogued?).
2. **`search --web`** — local FTS first, then Brave video for YouTube URLs not in DB.
3. Keep pure **`web`** for general Search-plan use without catalogue coupling.
4. Optional **`discover --catalogue`** auto-ingests up to 5 new channel suggestions.

## Consequences
- Brave becomes the discovery layer; yt-dlp remains the ownership layer; SQLite FTS
  remains the offline search layer.
- Agents get a clear loop without inventing ad-hoc Brave query recipes.
