# ADR 0003: Cobra + Viper CLI/config

## Status
Accepted

## Context
Need subcommands, rich help, flags, env, and optional config file with clear precedence for agents and humans.

## Decision
Cobra for command tree; Viper for config merge (flags > env > file > defaults). XDG config/data paths.

## Consequences
- Consistent `--json` / format flags across commands.
- Config show command can dump effective settings with secrets redacted.
