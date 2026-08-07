package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

const schemaV1 = `
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
`

func migrate(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `PRAGMA journal_mode=WAL;`); err != nil {
		return fmt.Errorf("wal: %w", err)
	}
	if _, err := db.ExecContext(ctx, `PRAGMA foreign_keys=ON;`); err != nil {
		return fmt.Errorf("foreign_keys: %w", err)
	}
	if _, err := db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS schema_migrations (
    version INTEGER PRIMARY KEY,
    applied_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);`); err != nil {
		return err
	}
	var ver int
	err := db.QueryRowContext(ctx, `SELECT COALESCE(MAX(version),0) FROM schema_migrations`).Scan(&ver)
	if err != nil {
		return err
	}
	if ver < 1 {
		if _, err := db.ExecContext(ctx, schemaV1); err != nil {
			return fmt.Errorf("apply schema v1: %w", err)
		}
		if _, err := db.ExecContext(ctx, `INSERT INTO schema_migrations(version) VALUES (1)`); err != nil {
			return err
		}
		ver = 1
	}
	if ver < 2 {
		// Richer metadata: engagement, languages, captions, fetched_at
		for _, stmt := range []string{
			`ALTER TABLE videos ADD COLUMN comment_count INTEGER`,
			`ALTER TABLE videos ADD COLUMN language TEXT`,
			`ALTER TABLE videos ADD COLUMN languages TEXT`,
			`ALTER TABLE videos ADD COLUMN has_subtitles INTEGER NOT NULL DEFAULT 0`,
			`ALTER TABLE videos ADD COLUMN has_auto_captions INTEGER NOT NULL DEFAULT 0`,
			`ALTER TABLE videos ADD COLUMN has_transcript INTEGER NOT NULL DEFAULT 0`,
			`ALTER TABLE videos ADD COLUMN fetched_at TEXT`,
			// backfill fetched_at from updated_at for existing rows
			`UPDATE videos SET fetched_at = updated_at WHERE fetched_at IS NULL OR fetched_at = ''`,
		} {
			if _, err := db.ExecContext(ctx, stmt); err != nil {
				// SQLite has no IF NOT EXISTS for ADD COLUMN; ignore duplicate column
				if !strings.Contains(err.Error(), "duplicate column") {
					return fmt.Errorf("apply schema v2 (%s): %w", stmt, err)
				}
			}
		}
		if _, err := db.ExecContext(ctx, `INSERT INTO schema_migrations(version) VALUES (2)`); err != nil {
			return err
		}
	}
	return nil
}
