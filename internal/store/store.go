// Package store provides SQLite persistence with FTS5 search.
package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// Store wraps a SQLite database.
type Store struct {
	db   *sql.DB
	path string
}

// Open opens (or creates) a SQLite database at path and runs migrations.
func Open(ctx context.Context, path string) (*Store, error) {
	// modernc DSN: path with query params
	dsn := path
	if path != ":memory:" {
		dsn = path + "?_pragma=foreign_keys(1)"
	}
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1) // SQLite write safety for simple CLI
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	s := &Store{db: db, path: path}
	if err := migrate(ctx, db); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

// Close closes the database.
func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

// Path returns the database path.
func (s *Store) Path() string { return s.path }

// DB exposes the underlying *sql.DB for advanced use.
func (s *Store) DB() *sql.DB { return s.db }

// UpsertChannel inserts or updates a channel.
func (s *Store) UpsertChannel(ctx context.Context, ch Channel) error {
	_, err := s.db.ExecContext(ctx, `
INSERT INTO channels (id, title, description, custom_url, url, thumbnail_url, last_catalogued, video_count, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, strftime('%Y-%m-%dT%H:%M:%SZ','now'), strftime('%Y-%m-%dT%H:%M:%SZ','now'))
ON CONFLICT(id) DO UPDATE SET
  title=excluded.title,
  description=COALESCE(excluded.description, channels.description),
  custom_url=COALESCE(excluded.custom_url, channels.custom_url),
  url=COALESCE(excluded.url, channels.url),
  thumbnail_url=COALESCE(excluded.thumbnail_url, channels.thumbnail_url),
  last_catalogued=COALESCE(excluded.last_catalogued, channels.last_catalogued),
  video_count=CASE WHEN excluded.video_count > 0 THEN excluded.video_count ELSE channels.video_count END,
  updated_at=strftime('%Y-%m-%dT%H:%M:%SZ','now')
`, ch.ID, ch.Title, nullStr(ch.Description), nullStr(ch.CustomURL), nullStr(ch.URL),
		nullStr(ch.ThumbnailURL), nullStr(ch.LastCatalogued), ch.VideoCount)
	return err
}

// UpsertVideo inserts or updates a video.
func (s *Store) UpsertVideo(ctx context.Context, v Video) error {
	_, err := s.db.ExecContext(ctx, `
INSERT INTO videos (id, channel_id, title, description, upload_date, duration, view_count, like_count, thumbnail_url, webpage_url, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, strftime('%Y-%m-%dT%H:%M:%SZ','now'), strftime('%Y-%m-%dT%H:%M:%SZ','now'))
ON CONFLICT(id) DO UPDATE SET
  channel_id=excluded.channel_id,
  title=excluded.title,
  description=COALESCE(NULLIF(excluded.description,''), videos.description),
  upload_date=COALESCE(NULLIF(excluded.upload_date,''), videos.upload_date),
  duration=COALESCE(excluded.duration, videos.duration),
  view_count=COALESCE(excluded.view_count, videos.view_count),
  like_count=COALESCE(excluded.like_count, videos.like_count),
  thumbnail_url=COALESCE(NULLIF(excluded.thumbnail_url,''), videos.thumbnail_url),
  webpage_url=COALESCE(NULLIF(excluded.webpage_url,''), videos.webpage_url),
  updated_at=strftime('%Y-%m-%dT%H:%M:%SZ','now')
`, v.ID, v.ChannelID, v.Title, nullStr(v.Description), nullStr(v.UploadDate),
		nullInt(v.Duration), nullInt64(v.ViewCount), nullInt64(v.LikeCount),
		nullStr(v.ThumbnailURL), nullStr(v.WebpageURL))
	return err
}

// TouchChannelCatalogued updates last_catalogued and video_count.
func (s *Store) TouchChannelCatalogued(ctx context.Context, channelID string, when time.Time) error {
	ts := when.UTC().Format(time.RFC3339)
	_, err := s.db.ExecContext(ctx, `
UPDATE channels SET
  last_catalogued = ?,
  video_count = (SELECT COUNT(*) FROM videos WHERE channel_id = ?),
  updated_at = strftime('%Y-%m-%dT%H:%M:%SZ','now')
WHERE id = ?
`, ts, channelID, channelID)
	return err
}

// GetChannel returns a channel by id or custom_url handle.
func (s *Store) GetChannel(ctx context.Context, idOrHandle string) (*Channel, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT id, title, COALESCE(description,''), COALESCE(custom_url,''), COALESCE(url,''),
       COALESCE(thumbnail_url,''), COALESCE(last_catalogued,''), video_count,
       created_at, updated_at
FROM channels
WHERE id = ? OR custom_url = ? OR custom_url = ? OR url LIKE ?
`, idOrHandle, idOrHandle, ensureAt(idOrHandle), "%"+idOrHandle+"%")
	var ch Channel
	err := row.Scan(&ch.ID, &ch.Title, &ch.Description, &ch.CustomURL, &ch.URL,
		&ch.ThumbnailURL, &ch.LastCatalogued, &ch.VideoCount, &ch.CreatedAt, &ch.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &ch, nil
}

// ListChannels returns all channels ordered by title.
func (s *Store) ListChannels(ctx context.Context) ([]Channel, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, title, COALESCE(description,''), COALESCE(custom_url,''), COALESCE(url,''),
       COALESCE(thumbnail_url,''), COALESCE(last_catalogued,''), video_count,
       created_at, updated_at
FROM channels ORDER BY title COLLATE NOCASE`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Channel
	for rows.Next() {
		var ch Channel
		if err := rows.Scan(&ch.ID, &ch.Title, &ch.Description, &ch.CustomURL, &ch.URL,
			&ch.ThumbnailURL, &ch.LastCatalogued, &ch.VideoCount, &ch.CreatedAt, &ch.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, ch)
	}
	return out, rows.Err()
}

// GetVideo returns a video by id.
func (s *Store) GetVideo(ctx context.Context, id string) (*Video, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT id, channel_id, title, COALESCE(description,''), COALESCE(upload_date,''),
       COALESCE(duration,0), view_count, like_count, COALESCE(thumbnail_url,''),
       COALESCE(webpage_url,''), created_at, updated_at
FROM videos WHERE id = ?`, id)
	var v Video
	var vc, lc sql.NullInt64
	err := row.Scan(&v.ID, &v.ChannelID, &v.Title, &v.Description, &v.UploadDate,
		&v.Duration, &vc, &lc, &v.ThumbnailURL, &v.WebpageURL, &v.CreatedAt, &v.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if vc.Valid {
		v.ViewCount = &vc.Int64
	}
	if lc.Valid {
		v.LikeCount = &lc.Int64
	}
	return &v, nil
}

// DeleteChannel removes a channel and its videos (cascade).
func (s *Store) DeleteChannel(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM channels WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("channel not found: %s", id)
	}
	return nil
}

// NewestUploadDate returns the newest upload_date for a channel (lexicographic).
func (s *Store) NewestUploadDate(ctx context.Context, channelID string) (string, error) {
	var d sql.NullString
	err := s.db.QueryRowContext(ctx, `
SELECT upload_date FROM videos
WHERE channel_id = ? AND upload_date IS NOT NULL AND upload_date != ''
ORDER BY upload_date DESC LIMIT 1`, channelID).Scan(&d)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return d.String, nil
}

// Stats returns high-level counts.
func (s *Store) Stats(ctx context.Context) (Stats, error) {
	var st Stats
	st.DBPath = s.path
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM channels`).Scan(&st.Channels)
	if err != nil {
		return st, err
	}
	err = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM videos`).Scan(&st.Videos)
	return st, err
}

// SearchVideos runs FTS and/or filters with pagination.
func (s *Store) SearchVideos(ctx context.Context, p SearchParams) ([]Video, error) {
	if p.Limit <= 0 {
		p.Limit = 20
	}
	if p.Offset < 0 {
		p.Offset = 0
	}
	after := normalizeDate(p.After)
	before := normalizeDate(p.Before)
	q := strings.TrimSpace(p.Query)

	var (
		rows *sql.Rows
		err  error
	)

	if q != "" {
		// FTS path
		match := buildFTSQuery(q)
		rows, err = s.db.QueryContext(ctx, `
SELECT v.id, v.channel_id, v.title, COALESCE(v.description,''), COALESCE(v.upload_date,''),
       COALESCE(v.duration,0), v.view_count, v.like_count, COALESCE(v.thumbnail_url,''),
       COALESCE(v.webpage_url,''), v.created_at, v.updated_at,
       bm25(videos_fts) AS score
FROM videos_fts
JOIN videos v ON v.rowid = videos_fts.rowid
WHERE videos_fts MATCH ?
  AND (? = '' OR v.channel_id = ?)
  AND (? = '' OR v.upload_date >= ?)
  AND (? = '' OR v.upload_date <= ?)
ORDER BY score
LIMIT ? OFFSET ?
`, match, p.ChannelID, p.ChannelID, after, after, before, before, p.Limit, p.Offset)
	} else {
		rows, err = s.db.QueryContext(ctx, `
SELECT v.id, v.channel_id, v.title, COALESCE(v.description,''), COALESCE(v.upload_date,''),
       COALESCE(v.duration,0), v.view_count, v.like_count, COALESCE(v.thumbnail_url,''),
       COALESCE(v.webpage_url,''), v.created_at, v.updated_at,
       0 AS score
FROM videos v
WHERE (? = '' OR v.channel_id = ?)
  AND (? = '' OR v.upload_date >= ?)
  AND (? = '' OR v.upload_date <= ?)
ORDER BY v.upload_date DESC, v.updated_at DESC
LIMIT ? OFFSET ?
`, p.ChannelID, p.ChannelID, after, after, before, before, p.Limit, p.Offset)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanVideos(rows)
}

func scanVideos(rows *sql.Rows) ([]Video, error) {
	var out []Video
	for rows.Next() {
		var v Video
		var vc, lc sql.NullInt64
		var score float64
		if err := rows.Scan(&v.ID, &v.ChannelID, &v.Title, &v.Description, &v.UploadDate,
			&v.Duration, &vc, &lc, &v.ThumbnailURL, &v.WebpageURL, &v.CreatedAt, &v.UpdatedAt, &score); err != nil {
			return nil, err
		}
		if vc.Valid {
			v.ViewCount = &vc.Int64
		}
		if lc.Valid {
			v.LikeCount = &lc.Int64
		}
		v.SearchScore = score
		out = append(out, v)
	}
	return out, rows.Err()
}

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func nullInt(n int) any {
	if n == 0 {
		return nil
	}
	return n
}

func nullInt64(p *int64) any {
	if p == nil {
		return nil
	}
	return *p
}

func ensureAt(s string) string {
	if s == "" {
		return s
	}
	if strings.HasPrefix(s, "@") {
		return s
	}
	if !strings.Contains(s, "/") && !strings.HasPrefix(s, "UC") {
		return "@" + s
	}
	return s
}

func normalizeDate(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	// YYYYMMDD -> YYYY-MM-DD
	if len(s) == 8 && !strings.Contains(s, "-") {
		return s[0:4] + "-" + s[4:6] + "-" + s[6:8]
	}
	return s
}

// buildFTSQuery turns a user query into a safe FTS5 MATCH expression.
func buildFTSQuery(q string) string {
	// Quote tokens for phrase-ish matching; join with AND.
	// For tokens containing punctuation (e.g. HTTP/2), also OR a stripped form
	// so "http2" can match titles that use "HTTP/2".
	parts := strings.Fields(q)
	if len(parts) == 0 {
		return `""`
	}
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.ReplaceAll(p, `"`, "")
		if p == "" {
			continue
		}
		stripped := strings.Map(func(r rune) rune {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
				return r
			}
			return -1
		}, p)
		if stripped != "" && stripped != p {
			out = append(out, `("`+p+`" OR "`+stripped+`")`)
		} else {
			out = append(out, `"`+p+`"`)
		}
	}
	if len(out) == 0 {
		return `""`
	}
	return strings.Join(out, " AND ")
}
