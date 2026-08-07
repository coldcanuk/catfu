# CLI framework (catfu)

**Choice:** `github.com/spf13/cobra` + `github.com/spf13/viper`

## Precedence (highest wins)

1. CLI flags
2. Environment variables (`CATFU_*`, plus `BRAVE_API_KEY`)
3. Config file (`$XDG_CONFIG_HOME/catfu/config.yaml` or `~/.config/catfu/config.yaml`)
4. Built-in defaults

## Global flags

| Flag | Env | Default |
|------|-----|---------|
| `--db` | `CATFU_DB` | `~/.local/share/catfu/catfu.db` (or XDG data) |
| `--config` | | optional path |
| `--json` | | false — agent-friendly machine output |
| `--format` | | `table` \| `json` \| `csv` |
| `--quiet` / `-q` | | reduce human noise |
| `--brave-api-key` | `BRAVE_API_KEY` | empty |
| `--log-level` | `CATFU_LOG_LEVEL` | `info` |

## Agent-friendly rules

- Every core command supports `--json`.
- Meaningful exit codes: 0 ok, 1 generic, 2 usage, 3 dependency missing, 4 not found.
- Progress on stderr when not `--json`/`--quiet`.
