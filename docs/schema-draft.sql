-- catfu schema draft (see internal/store for applied migrations)
PRAGMA journal_mode = WAL;
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS schema_migrations (
    version INTEGER PRIMARY KEY,
    applied_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE TABLE IF NOT EXISTS channels (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    custom_url TEXT,
    url TEXT,
    thumbnail_url TEXT,
    last_catalogued TEXT,
    video_count INTEGER DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE TABLE IF NOT EXISTS videos (
    id TEXT PRIMARY KEY,
    channel_id TEXT NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT,
    upload_date TEXT,
    duration INTEGER,
    view_count INTEGER,
    like_count INTEGER,
    thumbnail_url TEXT,
    webpage_url TEXT,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE INDEX IF NOT EXISTS idx_videos_channel_upload ON videos(channel_id, upload_date DESC);
CREATE INDEX IF NOT EXISTS idx_videos_upload_date ON videos(upload_date);

CREATE VIRTUAL TABLE IF NOT EXISTS videos_fts USING fts5(
    title,
    description,
    content='videos',
    content_rowid='rowid',
    tokenize='porter unicode61 remove_diacritics 2'
);

CREATE TRIGGER IF NOT EXISTS videos_ai AFTER INSERT ON videos BEGIN
  INSERT INTO videos_fts(rowid, title, description)
  VALUES (new.rowid, new.title, COALESCE(new.description, ''));
END;

CREATE TRIGGER IF NOT EXISTS videos_ad AFTER DELETE ON videos BEGIN
  INSERT INTO videos_fts(videos_fts, rowid, title, description)
  VALUES ('delete', old.rowid, old.title, COALESCE(old.description, ''));
END;

CREATE TRIGGER IF NOT EXISTS videos_au AFTER UPDATE ON videos BEGIN
  INSERT INTO videos_fts(videos_fts, rowid, title, description)
  VALUES ('delete', old.rowid, old.title, COALESCE(old.description, ''));
  INSERT INTO videos_fts(rowid, title, description)
  VALUES (new.rowid, new.title, COALESCE(new.description, ''));
END;
