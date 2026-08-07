// Package search provides high-level local catalogue search.
package search

import (
	"context"
	"time"

	"github.com/coldcanuk/catfu/internal/backends"
	"github.com/coldcanuk/catfu/internal/store"
)

// CatalogueSearcher searches the local SQLite catalogue.
type CatalogueSearcher struct {
	Store *store.Store
}

// Name implements backends.Searcher.
func (c *CatalogueSearcher) Name() string { return "catalogue" }

// Search implements backends.Searcher.
func (c *CatalogueSearcher) Search(ctx context.Context, q backends.SearchQuery) ([]backends.Result, error) {
	p := store.SearchParams{
		Query:     q.Query,
		ChannelID: q.ChannelID,
		Limit:     q.Limit,
		Offset:    q.Offset,
	}
	if q.After != nil {
		p.After = q.After.UTC().Format("2006-01-02")
	}
	if q.Before != nil {
		p.Before = q.Before.UTC().Format("2006-01-02")
	}
	videos, err := c.Store.SearchVideos(ctx, p)
	if err != nil {
		return nil, err
	}
	// channel titles map
	chTitles := map[string]string{}
	out := make([]backends.Result, 0, len(videos))
	for _, v := range videos {
		title := chTitles[v.ChannelID]
		if title == "" {
			if ch, _ := c.Store.GetChannel(ctx, v.ChannelID); ch != nil {
				title = ch.Title
				chTitles[v.ChannelID] = title
			}
		}
		url := v.WebpageURL
		if url == "" {
			url = "https://www.youtube.com/watch?v=" + v.ID
		}
		out = append(out, backends.Result{
			ID:          v.ID,
			Title:       v.Title,
			URL:         url,
			Description: v.Description,
			ChannelID:   v.ChannelID,
			Channel:     title,
			UploadDate:  v.UploadDate,
			Duration:    v.Duration,
			Source:      "catalogue",
			Kind:        "catalogue",
			Score:       v.SearchScore,
		})
	}
	return out, nil
}

// ParseDate parses YYYY-MM-DD or YYYYMMDD into time.Time (UTC midnight).
func ParseDate(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	layouts := []string{"2006-01-02", "20060102", time.RFC3339}
	var err error
	for _, l := range layouts {
		var t time.Time
		t, err = time.ParseInLocation(l, s, time.UTC)
		if err == nil {
			return &t, nil
		}
	}
	return nil, err
}

var _ backends.Searcher = (*CatalogueSearcher)(nil)
