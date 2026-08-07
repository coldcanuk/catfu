# yt-dlp subprocess streaming patterns (catfu)

## Pipeline

1. `exec.CommandContext(ctx, "yt-dlp", args...)`
2. `StdoutPipe()` → `bufio.Scanner` or `json.Decoder` for NDJSON lines
3. `Stderr` → logger (progress / warnings), never mix into result stream
4. On each JSON object: map fields → upsert
5. On ctx cancel: process killed via CommandContext

## NDJSON vs single JSON

- `--dump-json` with playlist → **one JSON object per line** (NDJSON stream). Use line reader + `json.Unmarshal`.
- `--dump-single-json` → one large object (channel resolve). Full decode OK.

## Partial failure

- Malformed line: log + continue (count errors).
- Non-zero exit after partial stream: return error wrapping count of ingested rows (caller may still commit successful upserts — we upsert per row in transaction batches).

## Progress

- Count lines processed; optional callback `OnVideo(n, id, title)`.
- Human: stderr progress; JSON mode: silent or progress events on stderr only.

## Timeouts

- No global hard timeout by default (large channels).
- Rely on user/agent cancellation + polite sleeps inside yt-dlp.
