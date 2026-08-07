# Credential / BYOK patterns (catfu)

## Principles

- Bring-your-own-key: catfu never embeds API keys.
- Never log secrets (API keys, tokens).
- `catfu config` and `catfu doctor` **redact** secret values (show `set` / `***`).

## Sources

| Secret / path | Env | Flag | Config file key |
|---------------|-----|------|-----------------|
| Brave API key | `BRAVE_API_KEY`, `CATFU_BRAVE_API_KEY` | `--brave-api-key` | `brave_api_key` |
| DB path | `CATFU_DB` | `--db` | `db` |
| Log level | `CATFU_LOG_LEVEL` | `--log-level` | `log_level` |
| yt-dlp path | `CATFU_YTDLP` | `--ytdlp` | `ytdlp` |

## Config file locations

1. `--config` path if set
2. `$XDG_CONFIG_HOME/catfu/config.yaml`
3. `~/.config/catfu/config.yaml`

## Data directory

Default DB: `$XDG_DATA_HOME/catfu/catfu.db` or `~/.local/share/catfu/catfu.db`.
