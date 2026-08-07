# ADR 0008: Brave credential storage (keychain + file fallback)

## Status
Accepted

## Context
Requiring users to `export BRAVE_API_KEY` every session is poor UX and encourages
shell history leakage. Config.yaml can store the key but is plain text and easy
to commit by mistake.

## Decision
1. Prefer **OS keychain** via `github.com/zalando/go-keyring` (service `catfu`).
2. Fallback: **0600 file** `$XDG_CONFIG_HOME/catfu/secrets` when keyring is unavailable.
3. CLI: `catfu auth set|status|clear`.
4. Precedence: flag > env > config.yaml > keyring/file.
5. Never print the raw token in status/doctor/config output.

## Consequences
- Interactive users get one-time setup.
- Headless/CI continues to use env/flags.
- Linux servers without Secret Service use the file fallback.
