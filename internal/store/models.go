package store

// Channel is a catalogued YouTube channel.
type Channel struct {
	ID             string `json:"id"`
	Title          string `json:"title"`
	Description    string `json:"description,omitempty"`
	CustomURL      string `json:"custom_url,omitempty"`
	URL            string `json:"url,omitempty"`
	ThumbnailURL   string `json:"thumbnail_url,omitempty"`
	LastCatalogued string `json:"last_catalogued,omitempty"`
	VideoCount     int    `json:"video_count"`
	CreatedAt      string `json:"created_at,omitempty"`
	UpdatedAt      string `json:"updated_at,omitempty"`
}

// Video is a catalogued YouTube video metadata row.
//
// Note: catfu never downloads media files. "Fetched" means metadata was
// retrieved via yt-dlp into the local SQLite catalogue.
type Video struct {
	ID           string `json:"id"`
	ChannelID    string `json:"channel_id"`
	Title        string `json:"title"`
	Description  string `json:"description,omitempty"`
	UploadDate   string `json:"upload_date,omitempty"`
	Duration     int    `json:"duration,omitempty"`
	ViewCount    *int64 `json:"view_count,omitempty"`
	LikeCount    *int64 `json:"like_count,omitempty"`
	CommentCount *int64 `json:"comment_count,omitempty"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
	WebpageURL   string `json:"webpage_url,omitempty"`

	// Language is the primary content language when yt-dlp reports it (e.g. "en").
	Language string `json:"language,omitempty"`
	// Languages is a JSON array of caption/subtitle language codes available
	// (manual + auto). Empty when unknown (common in flat-playlist mode).
	Languages string `json:"languages,omitempty"`
	// HasSubtitles is true when manual (human) subtitles exist.
	HasSubtitles bool `json:"has_subtitles"`
	// HasAutoCaptions is true when automatic captions exist (often used as transcript).
	HasAutoCaptions bool `json:"has_auto_captions"`
	// HasTranscript is true when either manual subs or auto-captions are available
	// (practical "can get a transcript" signal without downloading the text).
	HasTranscript bool `json:"has_transcript"`

	// FetchedAt is when this metadata row was last written from yt-dlp (ISO8601 UTC).
	// This is the "time we downloaded the data" — not media download.
	FetchedAt string `json:"fetched_at,omitempty"`

	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
	// SearchScore is set when returned from FTS.
	SearchScore float64 `json:"score,omitempty"`
}

// SearchParams filters catalogue search.
type SearchParams struct {
	Query     string
	ChannelID string
	After     string // YYYY-MM-DD or YYYYMMDD
	Before    string
	Limit     int
	Offset    int
}

// Stats summarises database contents.
type Stats struct {
	Channels int    `json:"channels"`
	Videos   int    `json:"videos"`
	DBPath   string `json:"db_path"`
}
