# ADR 0001: Pure-Go SQLite (modernc.org/sqlite)

## Status
Accepted

## Context
catfu needs local persistent storage with full-text search. CGO-based `mattn/go-sqlite3` complicates cross-compilation and CI.

## Decision
Use `modernc.org/sqlite` (pure Go, FTS5 supported). Go toolchain 1.25+ as required by current modernc releases.

## Consequences
- Easy `GOOS`/`GOARCH` cross-builds without a C compiler.
- Slightly different performance profile vs CGo SQLite; acceptable for CLI catalogue sizes.
- FTS5 external-content + triggers verified in research prototype.
