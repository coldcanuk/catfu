# ADR 0004: External yt-dlp + politeness

## Status
Accepted

## Context
YouTube metadata extraction is volatile; yt-dlp is the de-facto maintained extractor. Vendoring conflicts with GPLv3 packaging and update cadence.

## Decision
Require `yt-dlp` on PATH (configurable). Stream `--flat-playlist --dump-json --skip-download` with sleep flags. Optional `--full` for richer metadata. Never download media.

## Consequences
- `catfu doctor` must check yt-dlp.
- Users update yt-dlp independently.
- Rate-limit risk documented; sleeps configurable.
