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

// SearchKind selects a remote search vertical when supported.
type SearchKind string

const (
	SearchKindWeb   SearchKind = "web"
	SearchKindNews  SearchKind = "news"
	SearchKindVideo SearchKind = "video"
)

// SearchQuery is the portable search request for local and web backends.
type SearchQuery struct {
	Query      string
	ChannelID  string
	After      *time.Time
	Before     *time.Time
	Limit      int
	Offset     int
	Kind       SearchKind // web|news|video; empty => web for remote backends
	Country    string     // ISO 3166-1 alpha-2
	SearchLang string
	SafeSearch string // off|moderate|strict
	Freshness  string // pd|pw|pm|py or empty (use After/Before)
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
	Kind        string  `json:"kind,omitempty"` // web|news|video|catalogue
	Age         string  `json:"age,omitempty"`
	Score       float64 `json:"score,omitempty"`
}

// Searcher runs a query against a backend.
type Searcher interface {
	Name() string
	Search(ctx context.Context, q SearchQuery) ([]Result, error)
}
