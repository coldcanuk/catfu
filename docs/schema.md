# catfu database schema

SQLite database (WAL mode, foreign keys on). Pure-Go driver: `modernc.org/sqlite`.

## Tables

### `schema_migrations`

| Column | Type | Notes |
|--------|------|-------|
| version | INTEGER PK | migration id |
| applied_at | TEXT | ISO8601 UTC |

### `channels`

| Column | Type | Notes |
|--------|------|-------|
| id | TEXT PK | YouTube channel id (`UC…`) |
| title | TEXT | |
| description | TEXT | |
| custom_url | TEXT | e.g. `@handle` |
| url | TEXT | channel URL |
| thumbnail_url | TEXT | |
| last_catalogued | TEXT | ISO8601 when last full/partial catalogue finished |
| video_count | INTEGER | denormalised count |
| created_at / updated_at | TEXT | ISO8601 |

### `videos`

| Column | Type | Notes |
|--------|------|-------|
| id | TEXT PK | YouTube video id |
| channel_id | TEXT FK → channels | ON DELETE CASCADE |
| title | TEXT | |
| description | TEXT | often empty in flat catalogue mode |
| upload_date | TEXT | `YYYY-MM-DD` when known |
| duration | INTEGER | seconds |
| view_count / like_count | INTEGER | nullable |
| thumbnail_url / webpage_url | TEXT | |
| created_at / updated_at | TEXT | |

Indexes: `(channel_id, upload_date DESC)`, `(upload_date)`.

### `videos_fts` (FTS5)

External-content FTS5 over `videos(title, description)` with porter/unicode61 tokenizer. Kept in sync via INSERT/UPDATE/DELETE triggers.

## Search

```sql
SELECT v.* FROM videos_fts
JOIN videos v ON v.rowid = videos_fts.rowid
WHERE videos_fts MATCH ?
  AND optional channel_id / upload_date bounds
ORDER BY bm25(videos_fts)
LIMIT ? OFFSET ?;
```

Without a text query, results are filtered by channel/date and ordered by `upload_date DESC`.

See also `docs/schema-draft.sql` and `internal/store/migrate.go`.
