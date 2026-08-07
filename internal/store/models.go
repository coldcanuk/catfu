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
type Video struct {
	ID           string `json:"id"`
	ChannelID    string `json:"channel_id"`
	Title        string `json:"title"`
	Description  string `json:"description,omitempty"`
	UploadDate   string `json:"upload_date,omitempty"`
	Duration     int    `json:"duration,omitempty"`
	ViewCount    *int64 `json:"view_count,omitempty"`
	LikeCount    *int64 `json:"like_count,omitempty"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
	WebpageURL   string `json:"webpage_url,omitempty"`
	CreatedAt    string `json:"created_at,omitempty"`
	UpdatedAt    string `json:"updated_at,omitempty"`
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
	Channels int   `json:"channels"`
	Videos   int   `json:"videos"`
	DBPath   string `json:"db_path"`
}
