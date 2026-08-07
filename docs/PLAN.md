# catfu master plan (updated after Milestone 1.3)

## Concrete technology choices

| Area | Choice |
|------|--------|
| Module | `github.com/coldcanuk/catfu` |
| Language | Go 1.25+ (toolchain) |
| CLI | cobra + viper |
| SQLite | `modernc.org/sqlite` + FTS5 external-content |
| MCP | `github.com/modelcontextprotocol/go-sdk` stdio |
| Logging | `log/slog` |
| yt-dlp | external binary on PATH |
| License | GPLv3 |
| Web search | Brave Search API (BYOK thin client) |

## Scope unchanged

Open-source foundation: catalogue (no download), local SQLite+FTS, pluggable backends (Brave first), CLI with JSON everywhere, MCP stdio, docs.

## Revised notes from research

1. **Flat playlist** omits `upload_date`/`description` — default fast path; `--full` optional.
2. Channel resolve via `--dump-single-json` supplies UC id + handle.
3. Brave: `X-Subscription-Token`, max count 20, offset max 9.
4. FTS5 external-content + triggers verified with modernc.
5. MCP: typed `AddTool` + `StdioTransport`; log to stderr.

## Delivery phases

Phases 2–10 proceed as in the original plan; README remains Milestone 10.2.
