package search

import (
	"context"
	"strings"

	"github.com/coldcanuk/catfu/internal/backends"
	"github.com/coldcanuk/catfu/internal/backends/brave"
	"github.com/coldcanuk/catfu/internal/youtube"
)

// Hybrid combines local catalogue FTS with Brave YouTube video signals.
type Hybrid struct {
	Local *CatalogueSearcher
	Brave *brave.Client
}

// Search runs local first, then supplements with Brave video hits not already local.
// Results are tagged source=catalogue|brave-video.
func (h *Hybrid) Search(ctx context.Context, q backends.SearchQuery) ([]backends.Result, error) {
	var out []backends.Result
	seen := map[string]bool{}

	if h.Local != nil {
		local, err := h.Local.Search(ctx, q)
		if err != nil {
			return nil, err
		}
		for _, r := range local {
			key := r.ID
			if key == "" {
				key = r.URL
			}
			seen[key] = true
			out = append(out, r)
		}
	}

	if h.Brave == nil || strings.TrimSpace(q.Query) == "" {
		return out, nil
	}

	limit := q.Limit
	if limit <= 0 {
		limit = 20
	}
	// Fetch extra from Brave to fill after local
	need := limit
	if len(out) >= limit {
		// still add a few web supplements for discovery value
		need = limit + 5
	}

	bq := backends.SearchQuery{
		Query:      q.Query,
		Kind:       backends.SearchKindVideo,
		Limit:      minInt(need, 20),
		Country:    q.Country,
		SearchLang: q.SearchLang,
		SafeSearch: q.SafeSearch,
		Freshness:  q.Freshness,
		After:      q.After,
		Before:     q.Before,
	}
	remote, err := h.Brave.Search(ctx, bq)
	if err != nil {
		// Local results still valuable; surface partial
		if len(out) > 0 {
			return out, nil
		}
		return nil, err
	}
	for _, r := range remote {
		vid := youtube.VideoID(r.URL)
		if vid == "" {
			// keep non-YT video results only if host is youtube
			continue
		}
		if seen[vid] {
			continue
		}
		seen[vid] = true
		r.ID = vid
		r.URL = youtube.WatchURL(vid)
		if r.Kind == "" {
			r.Kind = "video"
		}
		if r.Source == "" {
			r.Source = "brave-video"
		}
		// Mark clearly as not yet catalogued
		if !strings.Contains(r.Source, "brave") {
			r.Source = "brave-video"
		}
		out = append(out, r)
		if len(out) >= need {
			break
		}
	}
	if len(out) > limit && limit > 0 {
		// Prefer keeping all local; trim remote overflow only if over 2x limit
		if len(out) > limit*2 {
			out = out[:limit*2]
		}
	}
	return out, nil
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
