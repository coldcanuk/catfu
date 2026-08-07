// Package backends defines pluggable catalogue and search interfaces.
package backends

import (
	"context"
	"time"
)

// CatalogueOpts controls channel metadata extraction.
type CatalogueOpts struct {
	FullMetadata bool
	Limit        int
	SleepRequest float64
	SleepMin     float64
	SleepMax     float64
	DateAfter    string
	Progress     func(n int, videoID, title string)
}

// Cataloguer ingests YouTube channel metadata into the local store.
type Cataloguer interface {
	CatalogueChannel(ctx context.Context, channelURL string, opts CatalogueOpts) error
}

// SearchQuery is the portable search request for local and web backends.
type SearchQuery struct {
	Query     string
	ChannelID string
	After     *time.Time
	Before    *time.Time
	Limit     int
	Offset    int
}

// Result is a normalised search hit.
type Result struct {
	ID          string  `json:"id,omitempty"`
	Title       string  `json:"title"`
	URL         string  `json:"url"`
	Description string  `json:"description,omitempty"`
	ChannelID   string  `json:"channel_id,omitempty"`
	Channel     string  `json:"channel,omitempty"`
	UploadDate  string  `json:"upload_date,omitempty"`
	Duration    int     `json:"duration,omitempty"`
	Source      string  `json:"source"`
	Score       float64 `json:"score,omitempty"`
}

// Searcher runs a query against a backend.
type Searcher interface {
	Name() string
	Search(ctx context.Context, q SearchQuery) ([]Result, error)
}
