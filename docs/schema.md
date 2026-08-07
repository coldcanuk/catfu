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
| last_catalogued | TEXT | ISO8601 when last catalogue finished |
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
| view_count / like_count / comment_count | INTEGER | nullable; need `--full` for reliable values |
| thumbnail_url / webpage_url | TEXT | |
| language | TEXT | primary content language (e.g. `en`) when reported |
| languages | TEXT | JSON array of caption/subtitle language codes |
| has_subtitles | INTEGER | 1 if manual (human) subtitles exist |
| has_auto_captions | INTEGER | 1 if automatic captions exist |
| has_transcript | INTEGER | 1 if manual **or** auto captions exist (practical transcript signal) |
| fetched_at | TEXT | ISO8601 when this metadata was last pulled from yt-dlp |
| created_at / updated_at | TEXT | row lifecycle |

Indexes: `(channel_id, upload_date DESC)`, `(upload_date)`.

**Schema version 2** adds engagement extras, language/caption fields, and `fetched_at`.

### `videos_fts` (FTS5)

External-content FTS5 over `videos(title, description)` with porter/unicode61 tokenizer. Kept in sync via INSERT/UPDATE/DELETE triggers.

## Flat vs full catalogue

| Mode | Command | What you get |
|------|---------|--------------|
| Fast (flat) | `catfu catalogue @ch --limit N` | ids, titles, sometimes duration; **usually no** views/likes/captions |
| Full | `catfu catalogue @ch --full --limit N` | views, likes, comments, language, subtitle/caption language list, descriptions |

catfu **never downloads video files or transcript text** — only metadata. `has_transcript` means “YouTube exposes captions you could fetch later,” not that the transcript body is stored.
