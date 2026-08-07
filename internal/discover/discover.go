// Package discover uses Brave Search to find YouTube channels/videos that
// can feed the local catalogue (force-multiplier over standalone web search).
package discover

import (
	"context"
	"fmt"
	"strings"

	"github.com/coldcanuk/catfu/internal/backends"
	"github.com/coldcanuk/catfu/internal/backends/brave"
	"github.com/coldcanuk/catfu/internal/store"
	"github.com/coldcanuk/catfu/internal/youtube"
)

// Options controls discovery.
type Options struct {
	Limit   int
	Country string
	// Kind: video (default), web, or both.
	Kind string
}

// ChannelHit is a channel suggested for cataloguing.
type ChannelHit struct {
	Handle      string   `json:"handle,omitempty"`
	ChannelID   string   `json:"channel_id,omitempty"`
	URL         string   `json:"url"`
	Catalogued  bool     `json:"catalogued"`
	LocalVideos int      `json:"local_videos,omitempty"`
	LocalTitle  string   `json:"local_title,omitempty"`
	Evidence    []string `json:"evidence,omitempty"` // sample result titles/urls
	Source      string   `json:"source"`
}

// VideoHit is a YouTube video found via Brave (may or may not be in catalogue).
type VideoHit struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
	InCatalogue bool   `json:"in_catalogue"`
	Age         string `json:"age,omitempty"`
	Source      string `json:"source"`
}

// Report is the discover command output.
type Report struct {
	Query    string       `json:"query"`
	Channels []ChannelHit `json:"channels"`
	Videos   []VideoHit   `json:"videos"`
	Note     string       `json:"note,omitempty"`
}

// Service discovers YouTube targets via Brave and enriches with local store state.
type Service struct {
	Brave *brave.Client
	Store *store.Store
}

// Discover runs Brave searches tuned for YouTube and correlates with the local DB.
func (s *Service) Discover(ctx context.Context, query string, opts Options) (*Report, error) {
	if s.Brave == nil {
		return nil, fmt.Errorf("brave client required")
	}
	q := strings.TrimSpace(query)
	if q == "" {
		return nil, fmt.Errorf("query is required")
	}
	limit := opts.Limit
	if limit <= 0 {
		limit = 15
	}
	kind := strings.ToLower(strings.TrimSpace(opts.Kind))
	if kind == "" {
		kind = "video"
	}

	rep := &Report{
		Query: q,
		Note:  "Brave finds public YouTube signals; catalogue them with: catfu catalogue <url-or-handle>",
	}

	var braveResults []backends.Result
	switch kind {
	case "web":
		rs, err := s.searchBrave(ctx, q+" site:youtube.com", backends.SearchKindWeb, limit, opts.Country)
		if err != nil {
			return nil, err
		}
		braveResults = rs
	case "both":
		rs1, err := s.searchBrave(ctx, q, backends.SearchKindVideo, limit, opts.Country)
		if err != nil {
			return nil, err
		}
		rs2, err := s.searchBrave(ctx, q+" site:youtube.com", backends.SearchKindWeb, min(limit, 10), opts.Country)
		if err != nil {
			return nil, err
		}
		braveResults = append(rs1, rs2...)
	default: // video
		// Prefer video vertical; bias query toward YouTube without excluding other hosts
		// (Brave video index is heavily YouTube-weighted). Also pull site:youtube.com web.
		rs1, err := s.searchBrave(ctx, q, backends.SearchKindVideo, limit, opts.Country)
		if err != nil {
			return nil, err
		}
		braveResults = rs1
		if len(braveResults) < limit {
			rs2, err := s.searchBrave(ctx, q+` site:youtube.com`, backends.SearchKindWeb, min(10, limit), opts.Country)
			if err == nil {
				braveResults = append(braveResults, rs2...)
			}
		}
	}

	chMap := map[string]*ChannelHit{}
	seenVid := map[string]bool{}

	for _, r := range braveResults {
		// Collect channel refs from URL + description + title
		blob := r.URL + " " + r.Description + " " + r.Title
		vids, chs := youtube.ExtractFromText(blob)
		if vid := youtube.VideoID(r.URL); vid != "" {
			vids = append(vids, vid)
		}
		// If URL is a channel page
		if ch := youtube.ParseChannelRef(r.URL); ch != nil {
			chs = append(chs, ch)
		}

		for _, vid := range vids {
			if seenVid[vid] {
				continue
			}
			seenVid[vid] = true
			inCat := false
			title := r.Title
			if s.Store != nil {
				if v, _ := s.Store.GetVideo(ctx, vid); v != nil {
					inCat = true
					if title == "" {
						title = v.Title
					}
				}
			}
			if title == "" {
				title = vid
			}
			rep.Videos = append(rep.Videos, VideoHit{
				ID:          vid,
				Title:       title,
				URL:         youtube.WatchURL(vid),
				Description: r.Description,
				InCatalogue: inCat,
				Age:         r.Age,
				Source:      r.Source,
			})
		}

		for _, ch := range chs {
			key := ch.ChannelID
			if key == "" {
				key = strings.ToLower(ch.Handle)
			}
			if key == "" {
				key = ch.URL
			}
			hit, ok := chMap[key]
			if !ok {
				hit = &ChannelHit{
					Handle:    ch.Handle,
					ChannelID: ch.ChannelID,
					URL:       ch.URL,
					Source:    "brave",
				}
				chMap[key] = hit
			}
			ev := r.Title
			if ev == "" {
				ev = r.URL
			}
			if len(hit.Evidence) < 5 {
				hit.Evidence = append(hit.Evidence, ev)
			}
		}
	}

	// Correlate channels with local store
	for _, hit := range chMap {
		if s.Store != nil {
			idOr := hit.ChannelID
			if idOr == "" {
				idOr = hit.Handle
			}
			if idOr != "" {
				if ch, _ := s.Store.GetChannel(ctx, idOr); ch != nil {
					hit.Catalogued = true
					hit.ChannelID = ch.ID
					hit.LocalTitle = ch.Title
					hit.LocalVideos = ch.VideoCount
					if hit.Handle == "" {
						hit.Handle = ch.CustomURL
					}
				}
			}
		}
		rep.Channels = append(rep.Channels, *hit)
	}

	// Cap video list
	if len(rep.Videos) > limit*2 {
		rep.Videos = rep.Videos[:limit*2]
	}
	return rep, nil
}

func (s *Service) searchBrave(ctx context.Context, q string, kind backends.SearchKind, limit int, country string) ([]backends.Result, error) {
	return s.Brave.Search(ctx, backends.SearchQuery{
		Query:   q,
		Kind:    kind,
		Limit:   limit,
		Country: country,
	})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
