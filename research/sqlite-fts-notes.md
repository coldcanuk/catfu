# SQLite FTS5 evaluation (catfu)

**Driver:** pure-Go `modernc.org/sqlite` (no CGO, FTS5 available, easy cross-compile).

## Schema approach

- Base tables: `channels`, `videos` (TEXT PKs = YouTube IDs).
- Implicit `rowid` retained (no `WITHOUT ROWID`) so external-content FTS can map `content_rowid='rowid'`.
- Virtual table:

```sql
CREATE VIRTUAL TABLE videos_fts USING fts5(
  title,
  description,
  content='videos',
  content_rowid='rowid',
  tokenize='porter unicode61 remove_diacritics 2'
);
```

## Triggers

Standard external-content pattern for INSERT/UPDATE/DELETE to keep FTS in sync. Application always mutates `videos` only; FTS is never written directly except rebuild.

## Date-range + FTS

```sql
SELECT v.*
FROM videos_fts f
JOIN videos v ON v.rowid = f.rowid
WHERE videos_fts MATCH ?
  AND (? IS NULL OR v.upload_date >= ?)
  AND (? IS NULL OR v.upload_date <= ?)
  AND (? IS NULL OR v.channel_id = ?)
ORDER BY bm25(videos_fts)
LIMIT ? OFFSET ?;
```

`upload_date` stored as `YYYYMMDD` text for lexicographic range compares, or ISO8601 `YYYY-MM-DD` — **decision: prefer `YYYY-MM-DD` ISO date** when known; empty string / NULL when unknown (flat mode).

## WAL + foreign keys

```sql
PRAGMA journal_mode=WAL;
PRAGMA foreign_keys=ON;
```

## Migrations

Simple `schema_version` table + embedded SQL steps (no heavy migration framework for v1).

## Prototype outcome

In-memory tests with modernc confirm FTS5 MATCH, triggers, and date filters work without CGO.
