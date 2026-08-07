# ADR 0005: Schema and FTS approach

## Status
Accepted

## Context
Need channel/video catalogue with full-text title/description search and date filters.

## Decision
SQLite tables `channels` + `videos` with TEXT YouTube IDs as PKs, WAL, foreign keys. External-content FTS5 `videos_fts` synced via triggers. Simple integer `schema_migrations` versioning.

## Consequences
- Upserts keep FTS consistent via triggers.
- Flat catalogue may leave description/upload_date empty until `--full` or future enrich.
