// Package youtube parses YouTube URLs and hostnames for discovery/enrichment.
package youtube

import (
	"net/url"
	"regexp"
	"strings"
)

var (
	reWatchID   = regexp.MustCompile(`(?i)(?:v=|/shorts/|/embed/|/live/|/v/|youtu\.be/)([A-Za-z0-9_-]{11})`)
	reChannelUC = regexp.MustCompile(`(?i)youtube\.com/channel/(UC[A-Za-z0-9_-]{20,})`)
	reHandle    = regexp.MustCompile(`(?i)youtube\.com/@([A-Za-z0-9._-]+)`)
	reUser      = regexp.MustCompile(`(?i)youtube\.com/(?:c|user)/([A-Za-z0-9._-]+)`)
)

// IsYouTubeHost reports whether host is a YouTube property.
func IsYouTubeHost(host string) bool {
	h := strings.ToLower(host)
	h = strings.TrimPrefix(h, "www.")
	h = strings.TrimPrefix(h, "m.")
	return h == "youtube.com" || h == "youtu.be" || h == "youtube-nocookie.com" || strings.HasSuffix(h, ".youtube.com")
}

// VideoID extracts an 11-char video id from a URL or bare id.
func VideoID(raw string) string {
	raw = strings.TrimSpace(raw)
	if re := regexp.MustCompile(`^[A-Za-z0-9_-]{11}$`); re.MatchString(raw) {
		return raw
	}
	if m := reWatchID.FindStringSubmatch(raw); len(m) == 2 {
		return m[1]
	}
	return ""
}

// ChannelRef is a channel identity hint extracted from text/URLs.
type ChannelRef struct {
	// Handle is @name when known.
	Handle string `json:"handle,omitempty"`
	// ChannelID is UCxxxx when known.
	ChannelID string `json:"channel_id,omitempty"`
	// URL is a catalogue-ready URL (@handle/videos or /channel/UC…/videos).
	URL string `json:"url,omitempty"`
	// Raw is the original matched URL when available.
	Raw string `json:"raw,omitempty"`
}

// ParseChannelRef extracts channel identity from a URL string.
func ParseChannelRef(raw string) *ChannelRef {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	// Bare @handle
	if strings.HasPrefix(raw, "@") && !strings.Contains(raw, "/") {
		h := raw
		return &ChannelRef{
			Handle: h,
			URL:    "https://www.youtube.com/" + h + "/videos",
			Raw:    raw,
		}
	}
	if m := reChannelUC.FindStringSubmatch(raw); len(m) == 2 {
		id := m[1]
		return &ChannelRef{
			ChannelID: id,
			URL:       "https://www.youtube.com/channel/" + id + "/videos",
			Raw:       raw,
		}
	}
	if m := reHandle.FindStringSubmatch(raw); len(m) == 2 {
		h := "@" + m[1]
		return &ChannelRef{
			Handle: h,
			URL:    "https://www.youtube.com/" + h + "/videos",
			Raw:    raw,
		}
	}
	if m := reUser.FindStringSubmatch(raw); len(m) == 2 {
		// /c/ or /user/ — still useful as catalogue target
		u, err := url.Parse(raw)
		if err == nil && u.Host != "" {
			path := strings.TrimSuffix(u.Path, "/")
			return &ChannelRef{
				URL: "https://www.youtube.com" + path + "/videos",
				Raw: raw,
			}
		}
	}
	return nil
}

// ExtractFromText finds YouTube video ids and channel refs in free text (URLs, descriptions).
func ExtractFromText(text string) (videoIDs []string, channels []*ChannelRef) {
	seenV := map[string]bool{}
	seenC := map[string]bool{}
	// Split on whitespace and common delimiters for URL-ish tokens
	fields := strings.FieldsFunc(text, func(r rune) bool {
		return r == ' ' || r == '\n' || r == '\t' || r == '<' || r == '>' || r == '"' || r == '\'' || r == ')' || r == '('
	})
	for _, f := range fields {
		f = strings.Trim(f, ".,;]")
		if vid := VideoID(f); vid != "" {
			if !seenV[vid] {
				seenV[vid] = true
				videoIDs = append(videoIDs, vid)
			}
		}
		if ch := ParseChannelRef(f); ch != nil {
			key := ch.ChannelID
			if key == "" {
				key = ch.Handle
			}
			if key == "" {
				key = ch.URL
			}
			if key != "" && !seenC[key] {
				seenC[key] = true
				channels = append(channels, ch)
			}
		}
	}
	return videoIDs, channels
}

// WatchURL builds a canonical watch URL.
func WatchURL(videoID string) string {
	return "https://www.youtube.com/watch?v=" + videoID
}
