# catfu

**catfu** is an open-source YouTube **channel metadata catalogue** (no video download).
It stores titles, descriptions, dates, and related fields in a local **SQLite + FTS5**
database, exposes an agent-friendly **CLI** (JSON everywhere), optional **Brave Search**,
and a **stdio MCP server**.

License: **GNU GPLv3**.  
Module: `github.com/coldcanuk/catfu`

## Features

- Catalogue channel metadata via external **yt-dlp** (streaming JSON, polite sleeps)
- Local full-text search (FTS5) with channel + date filters
- Pluggable search backends (local catalogue + Brave Web Search)
- CLI designed for humans **and** agents (`--json`, stable exit codes)
- MCP server (`catfu mcp`) for tool-using agents
- Pure-Go SQLite (`modernc.org/sqlite`) — easy cross-compile, no CGO

## Requirements

| Dependency | Notes |
|------------|--------|
| Go 1.25+ | build only |
| [yt-dlp](https://github.com/yt-dlp/yt-dlp) | **runtime**, must be on `PATH` (you install it) |
| Brave Search API key | optional, for `catfu web` / `web_search` tool |

yt-dlp is **not** bundled. Its license (Unlicense) is compatible with GPLv3; catfu only invokes it as an external process.

## Install

```bash
# yt-dlp (example)
curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o ~/.local/bin/yt-dlp
chmod a+rx ~/.local/bin/yt-dlp

# catfu
go install github.com/coldcanuk/catfu/cmd/catfu@latest
# or from source:
git clone https://github.com/coldcanuk/catfu.git
cd catfu && make build && ./bin/catfu doctor
```

## Quick start

```bash
catfu doctor --json
catfu catalogue @CTVNews --limit 100
catfu search "referendum" --json
catfu list
```

Default database: `$XDG_DATA_HOME/catfu/catfu.db` or `~/.local/share/catfu/catfu.db`.  
Override with `--db` / `CATFU_DB`.

## Configuration

Precedence: **flags > environment > config file > defaults**.

| Setting | Flag | Env | Config key |
|---------|------|-----|------------|
| Database path | `--db` | `CATFU_DB` | `db` |
| Brave API key | `--brave-api-key` | `BRAVE_API_KEY` | `brave_api_key` |
| yt-dlp binary | `--ytdlp` | `CATFU_YTDLP` | `ytdlp` |
| JSON output | `--json` | | |
| Output format | `--format` | | `table`/`json`/`csv` |
| Politeness sleeps | `--sleep-requests`, `--sleep-interval`, `--max-sleep-interval` | | |

Config file (optional): `$XDG_CONFIG_HOME/catfu/config.yaml`

```yaml
db: /path/to/catfu.db
brave_api_key: "…"   # prefer env in production
ytdlp: yt-dlp
log_level: info
```

`catfu config` prints effective settings with secrets redacted.

## Command reference

| Command | Purpose |
|---------|---------|
| `catalogue <channel>` | Ingest channel metadata (`--full`, `--limit`) |
| `search [query]` | Local FTS + `--channel` / `--after` / `--before` |
| `web` / `web-search` | Brave web search |
| `list` / `catalogues` | List channels |
| `status [channel]` | DB or channel status |
| `update <channel>` | Refresh / incremental catalogue |
| `info <id>` | Channel or video details |
| `delete <channel>` | Remove channel (`--force`) |
| `export` | Export results JSON/CSV |
| `version` | Version |
| `doctor` | Dependency & DB health |
| `config` | Effective config (redacted) |
| `completion` | Shell completions |
| `open <video-id>` | Open in browser |
| `stats` | Counts |
| `backends` | Backend list/status |
| `mcp` | Stdio MCP server |

Global: `--json`, `--format`, `--db`, `--quiet`, `--config`, sleep flags.

## Architecture

```
cmd/catfu          → CLI entry
internal/cli       → Cobra commands
internal/catalogue → yt-dlp wrapper + orchestration
internal/store     → SQLite schema, FTS, CRUD
internal/search    → local Searcher
internal/backends  → interfaces + Brave client
internal/mcp       → MCP stdio server
internal/config    → Viper/XDG config
```

See [docs/PLAN.md](docs/PLAN.md), [docs/schema.md](docs/schema.md), [docs/mcp-tools.md](docs/mcp-tools.md), and [docs/adr/](docs/adr/).

## Politeness / rate limits

YouTube may rate-limit aggressive metadata extraction. Defaults use yt-dlp sleep flags (`--sleep-requests 0.5`, interval 1–3s). Do not parallelise large catalogues from one IP without backoff. Temporary blocks are a known external risk.

## Development

```bash
make build
make test
make vet
```

## Contributing

1. Fork and branch from `main`
2. Keep milestones small; conventional commits preferred
3. `go test ./...` and `go vet ./...` must pass
4. Do not vendor yt-dlp; do not commit secrets or local `*.db`


## Community

- [Code of Conduct](CODE_OF_CONDUCT.md) — be cool, be kind, seek greatness
- [Contributing](CONTRIBUTING.md) — branch/PR workflow, review bar, and AI/agent norms

## License

Copyright (C) 2026 coldcanuk  
This program is free software under the **GNU General Public License v3** or later — see [LICENSE](LICENSE) and [NOTICE](NOTICE).
