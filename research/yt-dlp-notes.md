# yt-dlp research notes (catfu)

**Date:** 2026-08-07  
**yt-dlp version tested:** 2026.07.04  
**Sample target:** `https://www.youtube.com/@CTVNews/videos`

## Recommended invocation (metadata-only catalogue)

```bash
yt-dlp --flat-playlist --dump-json --skip-download \
  --sleep-requests 0.5 --sleep-interval 1 --max-sleep-interval 3 \
  --no-warnings \
  "$CHANNEL_URL"
```

Optional bounds:

- `--playlist-end N` — cap entries (tests / incremental slices)
- `--dateafter YYYYMMDD` — only works reliably without `--flat-playlist` (full extract)
- `--download-archive FILE` — resume pattern for download workflows; catfu uses DB state instead

Channel resolution (preferred first step):

```bash
yt-dlp --dump-single-json --skip-download --playlist-end 1 --no-warnings "$CHANNEL_URL"
```

Yields: `channel_id` (UC…), `channel`/`title`, `description`, `thumbnails`, `channel_follower_count`, tags.

## Flat-playlist field availability

| Field | Available in `--flat-playlist` | Notes |
|-------|--------------------------------|-------|
| `id` | yes | video id |
| `title` | yes | |
| `duration` | yes | seconds |
| `webpage_url` / `url` | yes | watch URL |
| `thumbnails` | yes | pick largest or default hq |
| `playlist_channel_id` | yes | UC… channel id |
| `playlist_channel` / `playlist_title` | yes | channel title |
| `playlist_uploader_id` | yes | `@handle` |
| `view_count` | often null | |
| `upload_date` / `timestamp` | often null | **missing in flat mode** |
| `description` | no | not in flat entries |
| `like_count` | no | |

**Implication:** default catalogue uses flat mode (fast, polite). Optional `--full` drops `--flat-playlist` for richer fields (upload_date, description, views) at higher latency / rate-limit risk.

## Channel URL normalisation

Accept and normalise to a videos tab URL when possible:

| Input form | Example | Notes |
|------------|---------|-------|
| `@handle` | `@CTVNews` | prefix `https://www.youtube.com/` |
| handle URL | `https://www.youtube.com/@CTVNews` | add `/videos` for listing |
| channel id | `UCxxxxxxxx` | `https://www.youtube.com/channel/UC…/videos` |
| `/channel/UC…` | full URL | ok |
| `/c/Name`, `/user/Name` | legacy | yt-dlp resolves |
| bare video URL | `watch?v=` | not a channel catalogue target |

## Shorts / Live / Premiere

- Flat entries may include Shorts mixed into `/videos` depending on tab.
- `live_status` often null in flat mode.
- Live / upcoming items: store if present; duration may be null.
- catfu does **not** download media — only metadata rows.

## Rate-limit / politeness strategy

Defaults (configurable via flags/env):

| Flag / setting | Default | Purpose |
|----------------|---------|---------|
| `--sleep-requests` | `0.5` | pause between HTTP requests |
| `--sleep-interval` | `1` | min sleep before each item |
| `--max-sleep-interval` | `3` | randomise upper bound |
| context cancel | always | agent/CLI interrupt |

Document risk: aggressive scraping can yield temporary IP blocks. Never parallelise multiple heavy channel dumps from the same host without backoff.

## Resume / incremental

1. Store `channels.last_catalogued` (ISO8601) and newest `videos.upload_date` when known.
2. On `update`, re-stream with optional playlist end / stop-when-seen-id of newest known video (application-level).
3. Prefer DB upsert by video `id` (idempotent) over yt-dlp archive files.

## External dependency policy

- **Never vendor** yt-dlp source or binary.
- Require `yt-dlp` on `PATH`.
- `catfu doctor` reports path + version.
- License: Unlicense → GPLv3 compatible (users install separately).
