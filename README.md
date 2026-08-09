# catfu

**catfu** is an open-source YouTube **channel metadata catalogue** (no video download).
It stores titles, descriptions, dates, and related fields in a local **SQLite + FTS5**
database, exposes an agent-friendly **CLI** (JSON everywhere), optional **Brave Search**,
a **stdio MCP server**, and **regex social-link extraction** from descriptions.

License: **GNU GPLv3**.  
Module: `github.com/coldcanuk/catfu`

## Features

- Catalogue channel metadata via external **yt-dlp** (streaming JSON, polite sleeps)
- Engagement + availability fields: views, likes, comments, languages, caption/transcript flags, `fetched_at` (use `--full`)
- Local full-text search (FTS5) with channel + date filters
- Pluggable search backends (local catalogue + Brave Search plan: web / news / video)
- **Brave × catalogue force multiplier**: `discover` finds channels to ingest; `search --web` merges local FTS with remote YouTube hits
- CLI designed for humans **and** agents (`--json`, stable exit codes)
- MCP server (`catfu mcp`) for tool-using agents
- Pure-Go SQLite (`modernc.org/sqlite`) — easy cross-compile, no CGO
- **Social link extraction** (`catfu socials`): regex scan of channel/video descriptions for X, Threads, Facebook, Instagram, Bluesky, Nostr, TikTok, LinkedIn, Mastodon, Discord, GitHub, GitLab, Telegram, WeChat, WhatsApp, LINE (see [docs/socials.md](docs/socials.md); prefer `catalogue --full` for descriptions)

## Requirements

| Dependency | Notes |
|------------|--------|
| Go 1.25+ | build only |
| [yt-dlp](https://github.com/yt-dlp/yt-dlp) | **runtime**, must be on `PATH` (you install it) |
| Brave **Search** plan API key | optional, for `web`, `discover`, `search --web` (not Answers) |

yt-dlp is **not** bundled. Its license (Unlicense) is compatible with GPLv3; catfu only invokes it as an external process.

## Install

### 1. yt-dlp (runtime dependency)

```bash
# example: install to ~/.local/bin (ensure that dir is on PATH)
curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o ~/.local/bin/yt-dlp
chmod a+rx ~/.local/bin/yt-dlp
# distro apt packages are often too old (YouTube returns 0 videos)
yt-dlp --version   # prefer 2025+ / 2026+
```

### 2. catfu (via `go install`)

Requires **Go 1.25+**.

```bash
go install github.com/coldcanuk/catfu/cmd/catfu@latest
```

**Where the binary lands:** `go install` does **not** write into the Go toolchain
dir (`/usr/local/go/bin`). It installs to:

| Condition | Install location |
|-----------|------------------|
| `GOBIN` set | `$GOBIN/catfu` |
| else (default) | `$GOPATH/bin/catfu` → usually **`~/go/bin/catfu`** |

Check:

```bash
go env GOBIN GOPATH
ls -la "$(go env GOPATH)/bin/catfu"
# or, if GOBIN is set:
# ls -la "$(go env GOBIN)/catfu"
```

### 3. Put `~/go/bin` on your `PATH` (if `which catfu` fails)

If install succeeded but `catfu` is not found, your shell cannot see `$GOPATH/bin`.

**bash** (`~/.bashrc`):

```bash
echo 'export PATH="$HOME/go/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

**zsh** (`~/.zshrc`):

```bash
echo 'export PATH="$HOME/go/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

**Current shell only** (temporary):

```bash
export PATH="$HOME/go/bin:$PATH"
```

Verify:

```bash
which catfu
catfu version
catfu doctor --json
```

You should see a path like `/home/you/go/bin/catfu` and JSON version/doctor output.

### 4. Or build from source

```bash
git clone https://github.com/coldcanuk/catfu.git
cd catfu && make build
./bin/catfu doctor
# optional: install into ~/go/bin
make install
```

## Quick start

```bash
catfu doctor --json
catfu catalogue @CTVNews --limit 100
catfu catalogue @CTVNews --full --limit 50   # richer metadata + descriptions
catfu search "referendum" --json
catfu list
catfu socials @CTVNews --json                 # needs descriptions (prefer --full)
```

Default database: `$XDG_DATA_HOME/catfu/catfu.db` or `~/.local/share/catfu/catfu.db`.  
Override with `--db` / `CATFU_DB`.


## Brave Search (optional BYOK)

catfu talks to Brave with **one Search plan subscription token** sent as
`X-Subscription-Token`.

| Plan | Needed? |
|------|---------|
| **Search** | Yes — Web, News, Video (and more) under one key |
| **Answers** | No — different product (AI answers); not used by catfu today |

Signup / credits: [Brave Search API](https://brave.com/search/api/) and
[dashboard](https://api-dashboard.search.brave.com/) (public pricing includes
~$5 free monthly credits on Search).

### Recommended: store the token once (`catfu auth`)

```bash
# Interactive (hidden input) — uses OS keychain when available
catfu auth set

# Or pipe / flag (less preferred; may land in shell history)
catfu auth set --token 'your-search-plan-token'
echo 'your-search-plan-token' | catfu auth set

catfu auth status   # never prints the secret
catfu web "golang concurrency" --country CA --json
catfu auth clear    # remove from keychain / secrets file
```

**Storage backends**

| Backend | When |
|---------|------|
| **OS keychain** | macOS Keychain, Windows Credential Manager, Linux Secret Service (GNOME Keyring / KWallet via libsecret) |
| **0600 secrets file** | Fallback if no keychain: `$XDG_CONFIG_HOME/catfu/secrets` |

### Precedence (highest first)

1. `--brave-api-key` flag (good for one-off / scripts)
2. `BRAVE_API_KEY` / `CATFU_BRAVE_API_KEY` env (good for CI/agents)
3. `brave_api_key` in config.yaml (discouraged for shared machines)
4. OS keychain / secrets file from `catfu auth set` (**best for daily use**)

```bash
# Agents / CI still work without keychain:
export BRAVE_API_KEY='…'
catfu web "query" --json
```

See [research/brave-api-notes.md](research/brave-api-notes.md) and
[docs/adr/0007-brave-search-plan.md](docs/adr/0007-brave-search-plan.md).


## Brave as a catalogue force multiplier

Brave and the local catalogue used to be **separate** tools. They now form a pipeline:

```text
discover (Brave) ──► catalogue (yt-dlp) ──► search (local FTS)
                         ▲
search --web ────────────┘  (also pulls remote YouTube hits)
```

| Command | Role |
|---------|------|
| `catfu discover "topic"` | Brave video/web → extract YouTube channels & videos; mark already-catalogued |
| `catfu catalogue @handle` | Own the metadata locally (yt-dlp) |
| `catfu search "q"` | Fast offline FTS over what you own |
| `catfu search "q" --web` | Local hits **plus** Brave YouTube videos not yet catalogued |
| `catfu web "q" --kind video` | Pure Brave (no catalogue correlation) |

Example agent loop:

```bash
catfu auth set   # once
catfu discover "golang concurrency patterns" --json
# pick handles/urls from channels[]
catfu catalogue @golang --limit 100
catfu search "goroutine" --json
catfu search "context cancel" --web --json   # local + remote YouTube
```

## Configuration

Precedence: **flags > environment > config file > defaults**.

| Setting | Flag | Env | Config key |
|---------|------|-----|------------|
| Database path | `--db` | `CATFU_DB` | `db` |
| Brave Search token | `--brave-api-key` | `BRAVE_API_KEY` / `CATFU_BRAVE_API_KEY` | `brave_api_key` |
| yt-dlp binary | `--ytdlp` | `CATFU_YTDLP` | `ytdlp` |
| JSON output | `--json` | | |
| Output format | `--format` | | `table`/`json`/`csv` |
| Log level | `--log-level` | | `log_level` |
| Politeness sleeps | `--sleep-requests`, `--sleep-interval`, `--max-sleep-interval` | | |

Config file (optional): `$XDG_CONFIG_HOME/catfu/config.yaml`

```yaml
db: /path/to/catfu.db
brave_api_key: "…"   # Search plan token; prefer env in production
ytdlp: yt-dlp
log_level: info
```

`catfu config` prints effective settings with secrets redacted.

## Command reference

| Command | Purpose |
|---------|---------|
| `catalogue <channel>` | Ingest channel metadata via yt-dlp (`--full`, `--limit`) |
| `update <channel>` | Incremental refresh of a catalogued channel |
| `search [query]` | Local FTS + `--channel` / date filters; `--web` adds Brave YouTube |
| `web` / `web-search` | Brave Search plan: `--kind` `web` / `news` / `video` |
| `discover <topic>` | Brave → YouTube channel/video suggestions to catalogue |
| `socials [channel]` | Extract social links/handles from descriptions (`--video`, `--platform`, `--json`) |
| `list` / `catalogues` | List catalogued channels |
| `status [channel]` | Database or channel catalogue status |
| `info <channel|video-id>` | Channel or video details |
| `delete <channel>` | Remove channel and its videos (`--force`) |
| `export` | Export search results or catalogue as JSON/CSV |
| `open <video-id>` | Open a video in the default browser |
| `stats` | High-level channel/video counts |
| `backends` | Search/catalogue backend list and status |
| `auth` | Manage stored Brave Search token (`set` / `status` / `clear`) |
| `config` | Effective config (secrets redacted) |
| `doctor` | Dependencies, DB accessibility, credentials |
| `mcp` | Stdio MCP server for tool-using agents |
| `completion` | Shell completions (bash/zsh/fish/powershell) |
| `version` | Print catfu version |

**Global flags:** `--json`, `--format` (`table` / `json` / `csv`), `--db`, `--config`, `--quiet`, `--log-level`, `--brave-api-key`, `--ytdlp`, `--sleep-requests`, `--sleep-interval`, `--max-sleep-interval`.

More examples: [docs/cli-examples.md](docs/cli-examples.md), [docs/socials.md](docs/socials.md).

## Architecture

```
cmd/catfu           → CLI entry
internal/cli        → Cobra commands
internal/catalogue  → yt-dlp wrapper + orchestration
internal/store      → SQLite schema, FTS, CRUD
internal/search     → local Searcher (+ hybrid with Brave)
internal/backends   → search interfaces + Brave client
internal/discover   → Brave → YouTube channel/video discovery
internal/social     → regex social link/handle extraction
internal/mcp        → MCP stdio server
internal/config     → Viper/XDG config
internal/secrets    → OS keychain / secrets-file credential store
internal/output     → table / JSON / CSV encoding
internal/youtube    → URL/handle normalization helpers
pkg/version         → build-time version string
```

See [docs/PLAN.md](docs/PLAN.md), [docs/schema.md](docs/schema.md), [docs/mcp-tools.md](docs/mcp-tools.md),
[docs/socials.md](docs/socials.md), [docs/cli-examples.md](docs/cli-examples.md), and [docs/adr/](docs/adr/).

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
- [Security policy](SECURITY.md) — how to report vulnerabilities (privately, please)

## License

Copyright (C) 2026 coldcanuk  
This program is free software under the **GNU General Public License v3** or later — see [LICENSE](LICENSE) and [NOTICE](NOTICE).
